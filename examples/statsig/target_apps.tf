resource "statsig_target_app" "test" {
  name        = "test_tf"
  description = "test target app created in terraform"
}

output "test_target_app" {
  value = statsig_target_app.test
}
