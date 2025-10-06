# Identity and metadata server workloads

A fake GCE-metadata server for on-premises workload identity binding

## Config

Configuration values can be passed via YAML or environment variables.
In the latter case, prefix with `AUTH_` and use `_` when entering a sub-section.
For example `host.clientKey` becomes `AUTH_HOST_CLIENTKEY`.

```yaml
# Port to listen on
port: 8080

# Project holding the workload identity pool
projectId: "trv-identity-server-testing"

# Number of the project referred to by projectId
projectNumber: "866597189115"

# Name of the workload idenityt pool
poolName: "kubernetes-pool"

# name of the workload identity provider
providerName: "production"

# Can be "kubernetes" or "host"
mode: "kubernetes"

token:
  lifetime:
    # Lifetime of access tokens
    access: '10m'
    # Lifetime of identity tokens
    identity: '10m'

cache:
  # Time to cache service account token for a pod ip.
  # If a pod IP is re-used within this timeframe, the wrong service
  # account might be returned. This setting is only used
  # if "mode" is set to "kubernetes".
  serviceAccountTTL: '2m'

  # Interval after which expired tokens are removed from memory
  # This is only required for tokens that are not frequently fetched.
  tokenCleanupInterval: '1h'

  # The minimum lifetime of a token before it is refreshed
  tokenMinLifetime: '1m'

# This section is only used when "mode" is set to "kubernetes"
kubernetes:
  # URL of the kubelet, used to resolve pods.
  # If this is set to an empty string, the kubernetes API is used
  # To retrieve the pod.
  # The kublet lookup requires "get" access to "nodes/proxy", while
  # The kubernetes API requires "get,list" access to "pods". Both
  # rights have to be set on cluster level.
  kubeletHost: "https://127.0.0.1:10250"

  # Path of the kubernetes CA path. The CA is loaded during startup
  # And required to talk to the kubelet.
  # This value only needs to be changed if the kubelet certificate
  # Is not signed by the same certificate as the kubernetes control plane.
  kubeletCaPath: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

# This section is only used when "mode" is set to "host"
host:
  # The URL of the identity server
  identityServer: "https://identity-server:443"

  # Path to the client certificate
  clientCert: "/etc/certs/machine/identity.pem"

  # Path to the client certificate private key
  clientKey: "/etc/certs/machine/identity.key"

  # Path the identity server CA cert.
  # If the ca cert is installed system-wide, this option can be
  # left empty
  cacert: ""

  # If a client certificate is less than this value, it will be
  # rotated.
  clientCertMinimumLifetime: 240h

  # Interval in which to check client certificate expiration
  clientCertRefresh: 24h
```

## Nix setup

In order to use this repository with nix, make sure that you have `direnv` and `just` installed globally.
Run `just init-nix` to activate direnv based nix for this repository.

## Testing

It is recommended to use unit-tests for debugging, as these will run without
the need for kubernetes or a valid client certificate.

> [!NOTE]
> Please note that the metadata-server assumes the existance of a registered
> kubernetes control plane when run in `kubernetes` mode. As of that, local
> testing should use `host` mode.

To test locally you first need a valid client certificate:

- Allow `127.0.0.1` and `::1`
- Use your machine name as identity
- Bind to `integration-test@trv-identity-server-testing.iam.gserviceaccount.com`

Now start an `identity-server` on your local machine as described in the
`identity-server` readme.

You can now start the `metdata-server` using the following arguments:

- `host.identityServer` set to `https://localhost:8443`
- `host.cacert` pointing to the cacert file passed to the identity-server
- `mode` set to `host`

## Links and docs

- [How GKE binding works](https://cdelmonte.medium.com/gke-behind-the-scenes-integrating-kubernetes-service-accounts-with-google-cloud-platform-a6661e6a6b7c)
- [Kubernetes service account tokens](https://medium.com/@radharamadoss/kubernetes-service-account-tokens-security-improvements-a9734c186afb)
- [Short lived credentials for service accounts](https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#rest_2)
- [JWT decoder](https://jwt.io)
- [Spiffe GCP proxy](https://github.com/GoogleCloudPlatform/professional-services/blob/d3e80783706f594aa393783f2136aa859fb1f1b1/tools/spiffe-gcp-proxy/main.go#L68)
- [Kube Google IAM](https://github.com/kernelpayments/kube-google-iam/blob/master/server/server.go#L246)
- [GitHub actions Google auth](https://github.com/google-github-actions/auth/blob/main/src/client/workload_identity_federation.ts)
- [Google Cloud SDK resolve credentials](https://github.com/googleapis/google-cloud-go/blob/bbb17a96aa8b49744e7b2d81403dccdce744520e/compute/metadata/metadata.go#L128)
- [Universe resolver](https://github.com/google-github-actions/actions-utils/blob/17346c80a1ccf50622eaf2848a9688b43de679a4/src/universe.ts#L24)
- [Google Cloud instance metadata endpoints](https://cloud.google.com/compute/docs/metadata/predefined-metadata-keys#instance-metadata)
- Cluster wide `metadata.google.internal` [explained here](https://stackoverflow.com/a/65338650)
- Eposing a daemonset [via HostIP](https://stackoverflow.com/questions/50216813/exposing-a-daemonset-service-for-consumption-by-pods-on-the-same-node/50221453#50221453)
- IP tables for a node can be modified [through a daemonset on host network](https://github.com/bowei/k8s-custom-iptables)
- [NodeJS "Am I on GCP" detection](https://github.com/googleapis/gcp-metadata/blob/main/src/index.ts#L447)
