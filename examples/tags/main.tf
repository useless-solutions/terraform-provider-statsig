terraform {
  required_providers {
    statsig = {
      # The first part of the source name is the "namespace" of the provider, while the second part is the actual provider name.
      source = "tbd/statsig"
    }
  }
}

provider "statsig" {
  console_api_key = "console-*"
}

data "statsig_tags" "all" {
  depends_on = [statsig_tag.test]
}

output "all_tags" {
  value = data.statsig_tags.all.tags
}

resource "statsig_tag" "test" {
  name        = "test_tf"
  description = "test tag created in terraform"
  is_core     = false
}

output "test_tag" {
  value = statsig_tag.test
}
