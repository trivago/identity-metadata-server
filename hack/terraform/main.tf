locals {
  project_id          = "trv-identity-server-testing"
  region              = "europe-west1"
  organization        = "trivago"
  organizational_unit = "Infrastructure Operations"
  country_code        = "DE"
}

// Enable required service APIs
// Please note that activation takes a few minutes,
// so be patient if the next steps fail.
resource "google_project_service" "services" {
  for_each = toset([
    "privateca.googleapis.com",
    "secretmanager.googleapis.com",
    "sts.googleapis.com",
    "iam.googleapis.com",
  ])

  project            = local.project_id
  service            = each.value
  disable_on_destroy = false
}

data "google_project" "handle" {
  project_id = local.project_id
}
