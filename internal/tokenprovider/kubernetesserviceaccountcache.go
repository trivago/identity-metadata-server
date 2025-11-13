package tokenprovider

import (
	"context"
	"identity-metadata-server/internal/shared"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/trivago/go-kubernetes/v4"
)

// KubernetesServiceAccountCache holds service account data already resolved from the kubernetes API.
// This avoids having to query the kubernetes API (multiple times) for every request.
type KubernetesServiceAccountCache struct {
	lock        *sync.Mutex
	data        map[string]kubernetesServiceAccountInfo
	ttl         time.Duration
	k8s         *kubernetes.Client
	hitMetric   prometheus.Counter
	missMetric  prometheus.Counter
	kubeletHost string
	apiMetrics  *shared.APIMetrics
}

// NewKubernetesServiceAccountCache creates a new service account cache with a given TTL
// for times stored inside the cache.
func NewKubernetesServiceAccountCache(client *kubernetes.Client, kubeletHost string, ttl time.Duration, apiMetrics *shared.APIMetrics) *KubernetesServiceAccountCache {
	hitMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metadata_server",
		Subsystem:   "serviceaccount_cache",
		Name:        "hits_total",
		Help:        "Total number of hits to the token cache.",
		ConstLabels: map[string]string{},
	})
	missMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metadata_server",
		Subsystem:   "serviceaccount_cache",
		Name:        "misses_total",
		Help:        "Total number of misses to the token cache.",
		ConstLabels: map[string]string{},
	})

	if err := shared.RegisterCollectorOrUseExisting(&hitMetric); err != nil {
		log.Warn().Err(err).Msg("Failed to register service account cache hit metric, metrics will not be available")
	}
	if err := shared.RegisterCollectorOrUseExisting(&missMetric); err != nil {
		log.Warn().Err(err).Msg("Failed to register service account cache miss metric, metrics will not be available")
	}

	return &KubernetesServiceAccountCache{
		lock:        &sync.Mutex{},
		data:        make(map[string]kubernetesServiceAccountInfo),
		ttl:         ttl,
		k8s:         client,
		hitMetric:   hitMetric,
		missMetric:  missMetric,
		kubeletHost: kubeletHost,
		apiMetrics:  apiMetrics,
	}
}

// Get returns the service account information for the given IP.
// If the service account is not known, it will be resolved from the
// kubernetes API.
// IPs will be re-resolved after serviceAccountCacheTTL has passed.
// This function is thread safe.
func (c *KubernetesServiceAccountCache) Get(podIP string, ctx context.Context) kubernetesServiceAccountInfo {
	c.lock.Lock()
	defer c.lock.Unlock()

	var (
		pod     kubernetes.NamedObject
		podList *KubeletPodList
	)

	if info, isKnown := c.data[podIP]; isKnown {
		// TODO: This is a bit dangerous for very shortlived pods.
		//       As far as I know, kubernetes pod IPs take some time to be recycled,
		//       not only because pods take some time to shot down/spin up, but this
		//       is not guaranteed.
		//       Because of this, the default TTL is quite short, so we essentially
		//       only cache the service account for requests that are coming in short
		//       succession.
		if time.Since(info.firstSeen) < c.ttl {
			c.hitMetric.Inc()
			return info
		}

		// We can extend the cache entry past TTL by verifying that the pod still exists
		// and is owned by the same service account. Best case this will result in 1
		// api call instead of 2 for every request after TTL.
		var err error
		if len(c.kubeletHost) > 0 {
			pod, podList, err = GetPodByIPviaKubelet(c.kubeletHost, podIP, 0, c.apiMetrics, ctx)
		} else {
			pod, err = GetPodByIPviaControlPlane(c.k8s, podIP, 0, c.apiMetrics, ctx)
		}

		if err == nil && pod != nil && info.IsOwnedBy(pod) {
			info.firstSeen = time.Now()
			c.data[podIP] = info
			c.hitMetric.Inc()
			return info
		}

		delete(c.data, podIP)
	}

	c.missMetric.Inc()

	// If kubeletHost is set, we will try to get the pod from the kubelet API.
	// Otherwise, we will try to get the pod from the control plane API.
	if len(c.kubeletHost) > 0 {
		return c.getFromKubelet(podIP, podList, ctx)
	}

	return c.getFromControlPlane(podIP, pod, ctx)
}

// getFromKubelet retrieves the service account information for the given pod IP
// by querying the kubelet API.
// If the pod or service account cannot be found, an empty kubernetesServiceAccountInfo
// will be returned.
func (c *KubernetesServiceAccountCache) getFromKubelet(podIP string, podList *KubeletPodList, ctx context.Context) kubernetesServiceAccountInfo {
	// The caller may pass a podlist from a previous call to avoid querying the kubelet twice.
	foundPods, err := GetAllPodsFromKubelet(c.kubeletHost, c.k8s, podList, c.apiMetrics, ctx)
	if err != nil {
		log.Error().Err(err).Str("podIP", podIP).Msg("Failed to get pods from kubelet")
		return kubernetesServiceAccountInfo{}
	}

	for foundPodIP, foundPodInfo := range foundPods {
		if cachedInfo, isKnown := c.data[foundPodIP]; isKnown {
			// Do not refresh up-to-date entries
			if cachedInfo.Equal(foundPodInfo) && time.Since(cachedInfo.firstSeen) < c.ttl {
				continue
			}
		}
		// Not known or outdated, store/update it
		c.data[foundPodIP] = foundPodInfo
	}

	// Clean up stale entries. As we have the full list of pods from the kubelet,
	// we can remove any entries that are not present anymore.
	for ip := range c.data {
		if _, foundOnNode := foundPods[ip]; !foundOnNode {
			delete(c.data, ip)
			continue
		}
	}

	if info, ok := c.data[podIP]; ok {
		return info
	}
	return kubernetesServiceAccountInfo{}
}

// getFromControlPlane retrieves the service account information for the given pod IP
// by querying the control plane API.
// If the pod or service account cannot be found, an empty kubernetesServiceAccountInfo
// will be returned.
func (c *KubernetesServiceAccountCache) getFromControlPlane(podIP string, pod kubernetes.NamedObject, ctx context.Context) kubernetesServiceAccountInfo {
	// The caller may pass a pod object from a previous call to avoid querying the control plane twice.
	if pod == nil {
		var err error
		pod, err = GetPodByIPviaControlPlane(c.k8s, podIP, 3, c.apiMetrics, ctx)
		if err != nil || pod == nil {
			log.Error().Err(err).Str("podIP", podIP).Msg("Failed to get pod for IP")
			return kubernetesServiceAccountInfo{}
		}
	}

	// Retrieve service account name from pod. If not found, default to "default", which
	// is the default kubernetes behavior.
	serviceAccountName, err := pod.GetString(kubernetes.Path{"spec", "serviceAccountName"})
	if err != nil {
		log.Warn().Err(err).Str("pod", pod.GetName()).Str("namespace", pod.GetNamespace()).Msg("No service account set")
		serviceAccountName = "default"
	}

	ksa := kubernetesServiceAccountInfo{
		firstSeen: time.Now(),
		owner:     pod,
		namespace: pod.GetNamespace(),
		name:      serviceAccountName,
	}

	requestStart := time.Now()
	statusCode := 200

	serviceAccount, err := c.k8s.GetNamespacedObject(kubernetes.ResourceServiceAccount, ksa.name, ksa.namespace, ctx)
	if err != nil {
		statusCode = -1
		log.Error().Err(err).
			Str("pod", pod.GetName()).
			Str("serviceAccount", ksa.name).
			Str("namespace", ksa.namespace).
			Msg("Failed to get service account for pod")
	} else {
		ksa.boundGSA, err = serviceAccount.GetAnnotation("iam.gke.io/gcp-service-account")
		if err != nil || ksa.boundGSA == "" {
			log.Error().Err(err).
				Str("serviceAccount", ksa.name).
				Str("namespace", ksa.namespace).
				Msg("Failed to get gcp service account annotation")
		}
	}

	_ = c.apiMetrics.TrackDuration(kubeAPIendpoint, metricPathServiceAccounts, time.Since(requestStart))
	_ = c.apiMetrics.TrackRequest(kubeAPIendpoint, metricPathServiceAccounts, statusCode)

	c.data[podIP] = ksa
	return ksa
}
