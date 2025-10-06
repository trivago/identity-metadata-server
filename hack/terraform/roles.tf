resource "google_project_iam_custom_role" "identity_server" {
  role_id     = "identity_server_integration_test"
  title       = "identity-server-integration-test"
  description = "Custom role containing permissions for the identity-server integration test"
  project     = local.project_id

  permissions = [
    "privateca.caPools.get",
    "privateca.certificateAuthorities.list",
    "privateca.certificateAuthorities.get",
    // For production use, the identity-server should have the create permission,
    // too. However, for integration tests, we only need to be able to list and get
    // "privateca.certificates.create",
  ]
}
