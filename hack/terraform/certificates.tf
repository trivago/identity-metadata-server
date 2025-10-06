resource "tls_private_key" "integration_test_server" {
  algorithm = "RSA"
  rsa_bits  = "4096"
}

resource "tls_private_key" "integration_test_client" {
  algorithm   = "ECDSA"
  ecdsa_curve = "P384"
}

resource "google_privateca_certificate" "integration_test_server" {
  project  = local.project_id
  location = local.region
  pool     = google_privateca_ca_pool.root_ca_pool.name

  certificate_authority = google_privateca_certificate_authority.root_ca.certificate_authority_id

  lifetime = "${10 * 365 * 24 * 3600}s" // valid for 10 years
  name     = "integration-test-server-rsa"

  config {
    subject_config {
      subject {
        common_name         = "identity-server"
        country_code        = local.country_code
        organization        = local.organization
        organizational_unit = local.organizational_unit
      }
      subject_alt_name {
        email_addresses = [google_service_account.identity_server.email]
        ip_addresses    = ["127.0.0.1", "::1"]
        dns_names       = ["identity-server", "localhost"]
      }
    }

    x509_config {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          digital_signature = true
          key_encipherment  = true
          data_encipherment = true
        }
        extended_key_usage {
          server_auth = true
        }
      }
    }

    public_key {
      format = "PEM"
      key    = base64encode(tls_private_key.integration_test_server.public_key_pem)
    }
  }

  lifecycle {
    ignore_changes = [lifetime]
  }
}

resource "google_privateca_certificate" "integration_test_client" {
  project  = local.project_id
  location = local.region
  pool     = google_privateca_ca_pool.root_ca_pool.name

  certificate_authority = google_privateca_certificate_authority.root_ca.certificate_authority_id

  lifetime = "${10 * 365 * 24 * 3600}s" // valid for 10 years
  name     = "integration-test-client-active"

  config {
    subject_config {
      subject {
        common_name         = "integration-test"
        country_code        = local.country_code
        organization        = local.organization
        organizational_unit = local.organizational_unit
      }
      subject_alt_name {
        email_addresses = [google_service_account.github.email]
        ip_addresses    = ["127.0.0.1", "::1"]
      }
    }

    x509_config {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          digital_signature = true
          key_encipherment  = true
          data_encipherment = true
        }
        extended_key_usage {
          client_auth = true
        }
      }
    }

    public_key {
      format = "PEM"
      key    = base64encode(tls_private_key.integration_test_client.public_key_pem)
    }
  }

  lifecycle {
    ignore_changes = [lifetime]
  }
}

resource "google_privateca_certificate" "integration_test_client_expired" {
  project  = local.project_id
  location = local.region
  pool     = google_privateca_ca_pool.root_ca_pool.name

  certificate_authority = google_privateca_certificate_authority.root_ca.certificate_authority_id

  lifetime = "60s" // valid for 1 minute (testing expired certificates)
  name     = "integration-test-client-revoked"

  config {
    subject_config {
      subject {
        common_name         = "integration-test"
        country_code        = local.country_code
        organization        = local.organization
        organizational_unit = local.organizational_unit
      }
      subject_alt_name {
        email_addresses = [google_service_account.github.email]
        ip_addresses    = ["127.0.0.1", "::1"]
      }
    }

    x509_config {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          digital_signature = true
          key_encipherment  = true
          data_encipherment = true
        }
        extended_key_usage {
          client_auth = true
        }
      }
    }

    public_key {
      format = "PEM"
      key    = base64encode(tls_private_key.integration_test_client.public_key_pem)
    }
  }

  lifecycle {
    ignore_changes = [lifetime]
  }
}

# !!! IMPORTANT !!!
# This resource needs to be revoked manually after creation
resource "google_privateca_certificate" "integration_test_client_revoked" {
  project  = local.project_id
  location = local.region
  pool     = google_privateca_ca_pool.root_ca_pool.name

  certificate_authority = google_privateca_certificate_authority.root_ca.certificate_authority_id

  lifetime = "${10 * 365 * 24 * 3600}s" // valid for 10 years
  name     = "integration-test-client-revoked2"

  config {
    subject_config {
      subject {
        common_name         = "integration-test"
        country_code        = local.country_code
        organization        = local.organization
        organizational_unit = local.organizational_unit
      }
      subject_alt_name {
        email_addresses = [google_service_account.github.email]
        ip_addresses    = ["127.0.0.1", "::1"]
      }
    }

    x509_config {
      ca_options {
        is_ca = true
      }
      key_usage {
        base_key_usage {
          digital_signature = true
          key_encipherment  = true
          data_encipherment = true
        }
        extended_key_usage {
          client_auth = true
        }
      }
    }

    public_key {
      format = "PEM"
      key    = base64encode(tls_private_key.integration_test_client.public_key_pem)
    }
  }

  lifecycle {
    ignore_changes = [lifetime]
  }
}
