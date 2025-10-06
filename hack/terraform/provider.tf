provider "google" {
  project               = local.project_id
  billing_project       = local.project_id
  user_project_override = true
}

provider "google-beta" {
  project               = local.project_id
  billing_project       = local.project_id
  user_project_override = true
}
