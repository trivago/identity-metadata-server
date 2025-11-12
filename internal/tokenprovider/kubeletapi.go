package tokenprovider

import (
	"context"
	"fmt"
	"identity-metadata-server/internal/shared"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	kubernetes "github.com/trivago/go-kubernetes/v4"
)

type KubeletPodInfo struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		UID       string `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		ServiceAccountName string `json:"serviceAccountName"`
	} `json:"spec"`
	Status struct {
		PodIP  string `json:"podIP"`
		HostIP string `json:"hostIP"`
		Phase  string `json:"phase"`
	} `json:"status"`
}

type KubeletPodList struct {
	Items []KubeletPodInfo `json:"items"`
}

var (
	// Used by GetServiceAccountToken to cache the service account token
	serviceAccountTokenGuard sync.Mutex
	// Used by GetServiceAccountToken to cache the service account token
	serviceAccountToken string
)

// NamedObject returns a kubernetes.NamedObject for KubeletPodInfo.
func (p *KubeletPodInfo) NamedObject() kubernetes.NamedObject {
	obj := kubernetes.NewNamedObject(p.Metadata.Name)

	_ = obj.SetNamespace(p.Metadata.Namespace)
	_ = obj.Set(kubernetes.Path{"kind"}, "Pod")
	_ = obj.Set(kubernetes.Path{"apiVersion"}, "v1")
	_ = obj.Set(kubernetes.Path{"metadata", "uid"}, p.Metadata.UID)
	_ = obj.Set(kubernetes.Path{"spec", "serviceAccountName"}, p.Spec.ServiceAccountName)
	_ = obj.Set(kubernetes.Path{"status", "podIP"}, p.Status.PodIP)
	_ = obj.Set(kubernetes.Path{"status", "hostIP"}, p.Status.HostIP)
	_ = obj.Set(kubernetes.Path{"status", "phase"}, p.Status.Phase)

	return obj
}

// GetServiceAccountToken retrieves the service account token from the mounted file in the pod.
// The token is cached for subsequent calls to avoid reading the file multiple times.
// This is fine, as these tokens never rotate.
// See https://kubernetes.io/docs/concepts/security/service-accounts/#get-a-token
// It is expected that the service account token is mounted at the default location
// "/var/run/secrets/kubernetes.io/serviceaccount/token".
func GetServiceAccountToken() (string, error) {
	serviceAccountTokenGuard.Lock()
	defer serviceAccountTokenGuard.Unlock()

	if len(serviceAccountToken) > 0 {
		// If we already have a token, return it immediately.
		return serviceAccountToken, nil
	}

	// The token has to be mounted to the pod via "automountServiceAccountToken: true" set
	// on the service account.
	// We can retrive the token from the default location where Kubernetes mounts it.
	const tokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	token, err := os.ReadFile(tokenFilePath)

	// We retry reading the token file a few times, as it might not be available immediately
	// after the pod is started.
	for i := 0; os.IsNotExist(err); i++ {
		if i > 5 {
			break
		}
		time.Sleep(100 * time.Millisecond)
		token, err = os.ReadFile(tokenFilePath)
	}

	if err == nil {
		serviceAccountToken = strings.TrimSpace(string(token))
		return serviceAccountToken, nil
	}
	return "", err
}

// GetPodsFromKubelet retrieves the list of pods from the kubelet API (pods on the current node).
// The service account needs to have a ClusterRole bound that allows "get" on "nodes/proxy".
// The kubelet API is typically available at https://localhost:10250 when the pod is running
// in host network mode.
func GetPodsFromKubelet(kubeletHost string, ctx context.Context) (*KubeletPodList, error) {
	token, err := GetServiceAccountToken()
	if err != nil {
		return nil, shared.WrapErrorf(err, "failed to read service account token from disk")
	}

	pods, err := shared.HttpGETJson[KubeletPodList](kubeletHost+"/pods", nil, map[string]string{
		"Accept":        "application/json",
		"User-Agent":    "metadata-server",
		"Authorization": "Bearer " + token,
	}, nil, 2, ctx)

	if err != nil {
		return nil, shared.WrapErrorf(err, "failed to get pods from kubelet API")
	}

	return pods, nil
}

// GetPodByIPviaKubelet retrieves the pod object for a given pod IP.
// If no pod is found, it will retry up to `retries` times with a
// linearly increasing delay.
func GetPodByIPviaKubelet(kubeletHost string, podIP string, retries int, ctx context.Context) (kubernetes.NamedObject, error) {
	if retries < 0 {
		return nil, fmt.Errorf("retries must be 0 or greater")
	}
	for tryCounter := 1; tryCounter <= retries+1; tryCounter++ {
		pods, err := GetPodsFromKubelet(kubeletHost, ctx)
		if err != nil {
			return nil, shared.WrapErrorf(err, "failed to get pods from kubelet API")
		}

		if len(pods.Items) == 0 {
			// As this process is running as a pod, something must be wrong with the kubelet API.
			return nil, fmt.Errorf("failed to get pods from kubelet API. This might be a permissions issue or the kubelet API is not reachable")
		}

		candidates := make([]*KubeletPodInfo, 0, len(pods.Items))
		for idx, pod := range pods.Items {
			// We also need to check "pending" pods to support InitContainers, too.
			if pod.Status.PodIP == podIP && (pod.Status.Phase == "Running" || pod.Status.Phase == "Pending") {
				// Note: We could early out here, but we need to properly react on
				// ambiguous lookups, i.e. multiple pods with the same IP.
				candidates = append(candidates, &(pods.Items[idx]))
			}
		}

		switch len(candidates) {
		case 0:
			log.Info().Msgf("No pod found for IP %s, retrying (%d/%d)...", podIP, tryCounter, retries)

			select {
			case <-ctx.Done():
				return nil, shared.WrapErrorf(ctx.Err(), "No pod found for IP %s", podIP)
			case <-time.After(200 * time.Duration(tryCounter) * time.Millisecond):
			}

			continue // Retry if no pod was found

		case 1:
			return candidates[0].NamedObject(), nil

		default:
			// Note: Resolving multiple pods for the same IP is tricky.
			// With the information available, we could only do this with a
			// ip:port to process ID matching and resolving the pod via containerd
			// using the cgroup name of the process. This requires access to the
			// host machine, i.e. containerd and /proc, which poses a major security
			// risk.

			// Special case: if the pod IP matches the host IP, it is using
			// host networking. This is ambiguous if more than one pod on this
			// node is using host networking, as they all share the same host IP.
			// We log this as a seprate error, as it's a known issue.
			if podIP == candidates[0].Status.HostIP {
				return nil, fmt.Errorf("multiple pods found using host networking. This is ambiguous and cannot be resolved properly")
			}

			// There should not be multiple pods with the same IP (except for host networking).
			// If there were, routing would not work properly, as the IP is used to route traffic to the pod.
			// As we test for the status phase, we should never end up here.
			return nil, fmt.Errorf("%d pods found for IP %s. This is ambiguous and cannot be resolved properly", len(candidates), podIP)
		}
	}

	return nil, fmt.Errorf("no pod found for IP %s after %d tries", podIP, retries)
}

// GetPodByIPviaControlPlane retrieves the pod object for a given pod IP.
// In contrast to GetPodFromIP, this function uses the Kubernetes control plane
// to retrieve the pod information. It does not require the kubelet API to be
// available and does not require the pod to be running on the same node as the
// metadata server.
// It will retry a few times in case the pod is not found yet.
func GetPodByIPviaControlPlane(client *kubernetes.Client, podIP string, retries int, ctx context.Context) (kubernetes.NamedObject, error) {
	if retries < 0 {
		return nil, fmt.Errorf("retries must be 0 or greater")
	}

	fieldSelector := fmt.Sprintf("status.podIP==%s,status.phase!=Succeeded,status.phase!=Failed,status.phase!=Unknown", podIP)
	for tryCounter := 1; tryCounter <= retries+1; tryCounter++ {

		candidates, err := client.ListAllObjects(kubernetes.ResourcePod, "", fieldSelector)
		if err != nil {
			return nil, shared.WrapErrorf(err, "failed to get pods from kubernetes API")
		}

		switch len(candidates) {
		case 0:
			log.Info().Msgf("No pod found for IP %s, retrying (%d/%d)...", podIP, tryCounter, retries)
			select {
			case <-ctx.Done():
				return nil, shared.WrapErrorf(ctx.Err(), "No pod found for IP %s", podIP)
			case <-time.After(200 * time.Duration(tryCounter) * time.Millisecond):
			}
			continue // Retry if no pod was found

		case 1:
			return candidates[0], nil // Found exactly one pod

		default:
			hostIP, _ := candidates[0].Get(kubernetes.Path{"status", "hostIP"})
			if podIP == hostIP {
				return nil, fmt.Errorf("multiple pods found using host networking. This is ambiguous and cannot be resolved properly")
			}
			return nil, fmt.Errorf("%d pods found for IP %s. This is ambiguous and cannot be resolved properly", len(candidates), podIP)
		}
	}

	return nil, fmt.Errorf("no pod found for IP %s after %d tries", podIP, retries)
}
