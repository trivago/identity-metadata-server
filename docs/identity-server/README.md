# identity-server

## Config

```yaml
port: 8443

server:
  # Path to the server's private key used for signing token requests.
  key: "server.pem"
  # Value if the `kid` field in the returned JWK
  keyName: "trivago-identity-server-01"
  # Value of the `iss` field in the generated JWT
  issuer: "https://identity-server"
  # The GCP service account this server is bound to
  identity: "identity-server@trv-identity-server-testing.iam.gserviceaccount.com"
  # The identity used in GCP IAM bindings for this instance. Defaults to hostname
  hostname: "identity-server"

  workloadIdentity:
    # Number of the GCP project holding the workload identity pool
    projectNumber: '866597189115'
    # Name of the workload identity pool
    PoolName: 'integration-test'
    # Name of the workload identity provider in the given pool
    providerName: 'identity-server'

  certAuthority:
    # ID of the project holding the certificate authority
    project: "trv-identity-server-testing"
    # Region of the certificate authority
    region: "europe-west1"
    # Name of the certificate authority pool
    poolName: "integration-test-ca-pool"
    # Name of the certificate authority
    name: "identity-server-ca"
    # The maximum time between CRL reloads
    crlRefresh: "24h"
    # The lifetime of certificates provided through the renew endpoint
    clientCertLifetime: "2160h"

  tls:
    # Path to the server certificate
    certificate: "/etc/certs/tls.crt"
    # Path to the server private key
    key: "/etc/certs/tls.key"
    # Time after which to check if the certificate or key has changed
    reload: "24h"
```

## Endpoints

| endpoint | method | authentication | description |
|----------|--------|----------------|-------------|
| `/jwks.json` | GET | none | returns the JWKS for server validation purposes |
| `/token` | GET | machine | get a signed token to identify the caller |
| `/identity` | GET | machine | get the service account assigned to the caller |
| `/refreshCrl` | POST | none | Refresh the CRL. Ratelimited to 1 request/min |
| `/healthz` | GET | none | Health check endpoint |
| `/readyz` | GET | none | Health check endpoint |

### Running an example server locally

The following steps need do be executed.
All steps are described below in more detail.

1. Get the keys/certificates
2. Build the code as described below
3. Run the server as described below

> [!WARNING]
> These steps require access to the GCP project.  
> You can replicate the project setup by applying the [terraform deployment](../../hack/terraform),
> with [local.project_id](../../hack/terraform/main.tf) set to a project you
> have access to. Adjust `GCP_PROJECT_ID` in the steps below to that value.

```shell
GCP_PROJECT_ID="trv-identity-server-testing"
CA_ROOT='mtls'

# Step 1: Setup certificates
CA_CERT_URL="$(gcloud privateca roots describe identity-server-ca --project="${GCP_PROJECT_ID}$" --pool=integration-test-ca-pool --location=europe-west1 --format='value(accessUrls.caCertificateAccessUrl)')"

mkdir -p "${CA_ROOT}/server"
curl -sL -o "${CA_ROOT}/cacert.pem" "${CA_CERT_URL}"

gcloud secrets versions access latest --project="${GCP_PROJECT_ID}$" --secret="integration-test-server-cert" > "${CA_ROOT}/server/server.cert"
gcloud secrets versions access latest --project="${GCP_PROJECT_ID}$" --secret="integration-test-server-key" > "${CA_ROOT}/server/server.key"

# Step 2: Build
go build -tags=jsoniter -ldflags="-s -w" -mod=mod ./cmd/identity-server

# Step 3: Run
./identity-server \
  --server.key="${CA_ROOT}/server/server.key" \
  --tls.certificate="${CA_ROOT}/server/server.cert" \
  --tls.key="${CA_ROOT}/server/server.key" \
  --server.issuer='https://identity-server' \
  --server.hostname='identity-server' \
  --server.identity="identity-server@${GCP_PROJECT_ID}$.iam.gserviceaccount.com" \
  --server.workloadidentity.providername='integration-test' \
  --server.workloadidentity.poolname='integration-test' \
  --server.certauthority.project="${GCP_PROJECT_ID}$" \
  --server.certauthority.region='europe-west1' \
  --server.certauthority.poolname='integration-test-ca-pool' \
  --server.certauthority.name='identity-server-ca'
```

### token request

> [!WARNING]
> Token requests require a valid client certificate and key.
> The certificate must be coming from the certificate authority, otherwise
> the handshake will fail.

```json
{
  "audience": "identity provider audience",
  "lifetime": "golang duration string (optional, defaults to 10m)"
}
```

#### token request curl example

This expects a [local server](#running-an-example-server-locally) to be running.  

```shell
CA_ROOT='mtls'

curl -vvv -X GET -H 'Content-Type: application/json' \
    --resolve 'identity-server:8443:127.0.0.1' \
    --cert "${CA_ROOT}/generated/client.cert" \
    --key "${CA_ROOT}/generated/client.key"  \
    --cacert "${CA_ROOT}/cacert.pem"  \
    -d "{\"audience\": \"test\", \"lifetime\": \"10m\"}" \
    'https://identity-server:8443/token'
```

## Concept

A client certifcate needs to be created at the certificate authority for each
host. The certificate must contain the service account as email as well as a
set of ip-addresses that are used for origin identitifaction.

All communication to the identity-server require HTTPS.
Endpoints related to client identity require a client mTLS certificate to be
presented.  
The identity-server only accepts requests from the local network.

### Identity request

An identity request uses a mTLS `identity-certificate` to verify the host.
By using mTLS we guard against the most common forms of attacks through
a proven and actively developed technology. We also use the metadata of the
required client certificate to identify the calling machine.

As a client certificate is prone to exfiltration, we take additional measures
to limit the blast radius of such a case, and to make it harder to use a
client certificate outside of the machine the certificate is assigned to.

An identity request does not take any parameters. The only identity that can
be assumed by the `host` is the one mentioned in the certificate used for
authentication.  
We also validate the host by matching the origin of a request to the IP
addresses stored in the certificate. We also check if the certificate is
expired or has been revoked, by utilizing the CRL provided by the certificate
autority.

> [!WARNING]
> We use IPs as a secondary identification method (next to the client
> certificate). This check is vulnerable to IP spoofing. However, an attacker
> would have to spoof both, the IP and the TLS communication and would still
> only gain access to _one_ identity.  
> We consider this measure to be "sufficient enough" for the time being.

### Additional considerations

In order to guard against exfiltration we would benefit from a "write-only",
hardware based key storage with the support for mTLS. This way an exfiltration
of the private key would become impossible.
