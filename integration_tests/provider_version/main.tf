terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "= 3.44"
    }
  }
  required_version = "~> 1.0"
}

provider "azurerm" {
  features {}
}
