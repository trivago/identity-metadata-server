

resource "google_service_account" "identity_server" {
  account_id   = "identity-server"
  display_name = "identity-server"
  project      = local.project_id
}

resource "google_service_account" "github" {
  account_id   = "github"
  display_name = "github"
  project      = local.project_id
}

resource "google_project_iam_member" "identity_server" {
  for_each = toset([
    "serviceAccount:github@${local.project_id}.iam.gserviceaccount.com",
    google_service_account.identity_server.member
  ])

  project = local.project_id

  role   = google_project_iam_custom_role.identity_server.name
  member = each.value
}
