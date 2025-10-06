

locals {
  secrets = {
    "integration-test-server-cert"    = google_privateca_certificate.integration_test_server.pem_certificate,
    "integration-test-server-key"     = tls_private_key.integration_test_server.private_key_pem_pkcs8,
    "integration-test-client-cert"    = google_privateca_certificate.integration_test_client_expired.pem_certificate,
    "integration-test-client-key"     = tls_private_key.integration_test_client.private_key_pem_pkcs8,
    "integration-test-client-revoked" = google_privateca_certificate.integration_test_client_revoked.pem_certificate,
  }
}

resource "google_secret_manager_secret" "secret" {
  for_each = local.secrets

  project   = local.project_id
  secret_id = each.key

  replication {
    user_managed {
      replicas {
        location = local.region
      }
    }
  }
}

resource "google_secret_manager_secret_iam_member" "access" {
  for_each = local.secrets
  project  = local.project_id

  secret_id = google_secret_manager_secret.secret[each.key].id

  role   = "roles/secretmanager.secretAccessor"
  member = google_service_account.github.member
}

resource "google_secret_manager_secret_version" "secret_data" {
  for_each    = local.secrets
  secret      = google_secret_manager_secret.secret[each.key].id
  secret_data = each.value
}

resource "google_secret_manager_secret_version" "client_cert_valid" {
  secret      = google_secret_manager_secret.secret["integration-test-client-cert"].id
  secret_data = google_privateca_certificate.integration_test_client.pem_certificate

  // Make sure the second version is created after the first one
  depends_on = [
    google_secret_manager_secret_version.secret_data["integration-test-client-cert"]
  ]
}
