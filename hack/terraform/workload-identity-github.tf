// These resources are required to run the integration tests from github actions.

resource "google_iam_workload_identity_pool" "github" {
  project = local.project_id

  workload_identity_pool_id = "github"
  display_name              = "GitHub binding pool"
  description               = "Identity pool storing GitHub bindings"

  depends_on = [
    google_project_service.services["iam.googleapis.com"]
  ]
}

resource "google_iam_workload_identity_pool_provider" "github_actions" {
  project = local.project_id

  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github-actions"
  display_name                       = "GitHub Actions"
  description                        = "GitHub Actions token trade using OIDC"

  attribute_condition = <<-COND
    attribute.repository_owner == "trivago"
  COND
  attribute_mapping = {
    "google.subject"             = "assertion.sub"
    "attribute.actor"            = "assertion.actor"
    "attribute.aud"              = "assertion.aud"
    "attribute.repo"             = "assertion.repository"
    "attribute.repository_owner" = "assertion.repository_owner"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

resource "google_service_account_iam_member" "binding" {
  service_account_id = google_service_account.github.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repo/${local.organization}/identity-metadata-server"
}
