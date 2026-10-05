terraform {
  required_version = ">= 1.10"

  # Credential-free providers: the test really applies and destroys this module, so it
  # must not need a cloud account.
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "~> 3.7"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}
