locals {
  // This is a manual copy, as there is no integration test server running
  // when applying this piece of terraform. So below on how this is intended
  // to work on production environments.
  jwks = <<-EOF
  {
    "keys": [
        {
            "e": "AQAB",
            "kid": "trivago-identity-server-01",
            "kty": "RSA",
            "n": "vfHcODF-1EN3epgwE8w8ph1-EH3Lf0Oe7_kcwKZ6FGt_u-J8tfi_zhlMUcNuqg_tKUftIcKr5xCwrncIkZPERivDFBHYZLH4ZSSEgA6Q3cTdbUBukM7etXrhXPp5hngAeDKphUh6XTvUWlS8uYN3jteL7wO6oxm9opMr2HIUrxNnKljPK-lMEK8ytTngqqJUmcXJg4UEgJuLbXlsDT1hLramD3WLmpYco_mXFqLcgd_3xozsndbMkU9zCjg4J2GaxvchuLDUYDypPQ-O3jCuuBx6yfvdhGdESsLGnzH_XImlxvjEn1tOr7uW7HyAC9Iqc7XPvSYPcr7o3QNClLe215ABfnYkB6c9S6a1C58s3mehCjqYXOFYg8gZS7MRZqPtN45Lvi16Wvjbcjc4kyL_jEJtZSEAIOWPcH3uuHwBRgg8vkImyOe7cRMscR8rB39mxFCb2a_sPqJDEuntIEvmsbSdNH1feC4UFoJqU6-9x3QruOg3r8bmHHWYZYSCVMiLsoSCk720d1w03NUuzLauKysYOghsqOsjPdcVIZDNAFv7i3vCiyqEpwHxSZAVmiR28YqMHME-38luqLTAK9O1cHDFmpJlOc8-igmlk2Hk7N0Si3S9rbs7bXX51RachHx8Lowr5sDiTabPv6uM5ZEn7A7pZVZYUkfRlOPYeCVIlgM",
            "use": "sig"
        }
    ]
  }
  EOF
}

// You can use a "http" resource to query the "/jwks.json" endoint of a
// running identity server for automatic retrieval.

// data "http" "jwks" {
//   url = "https://identity/jwks.json"
//   ca_cert_pem = data.google_secret_manager_secret_version.identity-root-ca.secret_data
//   request_headers = {
//     Accept = "application/jwk-set+json"
//   }
// }

resource "google_iam_workload_identity_pool" "integration-test" {
  project = local.project_id

  workload_identity_pool_id = "integration-test"
  display_name              = "machine identities"
  description               = "Identity pool storing bindings to integration-test machine identities"

  depends_on = [
    google_project_service.services["iam.googleapis.com"]
  ]
}

resource "google_iam_workload_identity_pool_provider" "integration_test" {
  project = local.project_id

  workload_identity_pool_id          = google_iam_workload_identity_pool.integration-test.workload_identity_pool_id
  workload_identity_pool_provider_id = "integration-test"
  display_name                       = "integration test"
  description                        = "integration-test token trade using OIDC"

  attribute_mapping = {
    "google.subject"                 = "assertion.sub"
    "attribute.service_account_name" = "assertion['node_claims']['identity']"
  }

  // This is important. We only allow the integration test identity to be available through this pool
  // The value must be equal to the value of `--server.hostname` passed to the identity server
  attribute_condition = <<-EOT
    google.subject == "identity-server"
  EOT

  oidc {
    issuer_uri = "https://identity-server"
    jwks_json  = local.jwks
  }
}

resource "google_service_account_iam_member" "identity_server_bindings" {
  service_account_id = google_service_account.identity_server.name
  role               = "roles/iam.workloadIdentityUser"

  member = "principal://iam.googleapis.com/projects/${data.google_project.handle.number}/locations/global/workloadIdentityPools/integration-test/subject/identity-server"
}
