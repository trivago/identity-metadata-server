terraform {
  required_version = "1.12.1"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.20.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "6.20.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.1.0"
    }
  }
}
