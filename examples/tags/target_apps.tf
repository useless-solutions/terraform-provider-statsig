data "statsig_target_apps" "all" {
  depends_on = [statsig_target_app.test]
}

output "all_tags" {
  value = data.statsig_target_apps.all.tags
}

resource "statsig_target_app" "test" {
  name        = "test_tf"
  description = "test target app created in terraform"
}

output "test_tag" {
  value = statsig_tag.test
}
