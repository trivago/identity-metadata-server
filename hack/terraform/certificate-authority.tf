resource "google_privateca_ca_pool" "root_ca_pool" {
  name     = "integration-test-ca-pool"
  project  = local.project_id
  location = local.region

  // The enterprise tier is required for listing, describing,
  // and revoking certificates
  tier = "ENTERPRISE"

  publishing_options {
    publish_ca_cert = true
    publish_crl     = true
  }

  depends_on = [
    google_project_service.services["privateca.googleapis.com"]
  ]
}

resource "google_privateca_certificate_authority" "root_ca" {
  pool                     = google_privateca_ca_pool.root_ca_pool.name
  certificate_authority_id = "identity-server-ca"
  project                  = local.project_id
  location                 = local.region
  config {
    subject_config {
      subject {
        common_name         = "identity-server-ca"
        organizational_unit = local.organizational_unit
        organization        = local.organization
        country_code        = local.country_code
      }
    }

    x509_config {
      ca_options {
        // is_ca *MUST* be true for certificate authorities
        is_ca = true
      }
      key_usage {
        base_key_usage {
          // cert_sign and crl_sign *MUST* be true for certificate authorities
          cert_sign = true
          crl_sign  = true
        }
        extended_key_usage {
        }
      }
    }
  }

  lifetime = "${10 * 365 * 24 * 3600}s" // valid for 10 years
  key_spec {
    algorithm = "EC_P256_SHA256"
  }

  // Disable CA deletion related safe checks for easier cleanup.
  deletion_protection                    = false
  skip_grace_period                      = true
  ignore_active_certificates_on_deletion = true
}

resource "google_privateca_ca_pool_iam_member" "pool_reader" {
  for_each = toset([
    google_service_account.github.member,
    google_service_account.identity_server.member
  ])

  project = local.project_id
  ca_pool = google_privateca_ca_pool.root_ca_pool.id

  role   = "roles/privateca.poolReader"
  member = each.value
}
