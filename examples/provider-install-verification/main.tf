terraform {
  required_providers {
    statsig = {
      # The first part of the source name is the "namespace" of the provider, while the second part is the actual provider name.
      source = "tbd/statsig"
    }
  }
}

provider "statsig" {
  console_api_key = "console-xxx"
}

data "statsig_tags" "example" {}
