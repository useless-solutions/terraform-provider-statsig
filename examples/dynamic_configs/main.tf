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

# data "statsig_dynamic_configs" "all" {
#   depends_on = [statsig_dynamic_config.test]
# }

# output "all_dynamic_configs" {
#   value = data.statsig_dynamic_configs.all.dynamic_configs
# }

resource "statsig_dynamic_config" "test" {
  name        = "test_tf"
  description = "test dynamic_config created in terraform"
  id_type     = "userID"
  rules       = []
  default_value = {
    "key1" : "value1",
    "key2" : "value2"
  }
}

output "test_dynamic_config" {
  value = statsig_dynamic_config.test
}
