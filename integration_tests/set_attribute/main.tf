terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "= 3.66"
    }
  }
  required_version = "~> 1.0"
}

provider "azurerm" {
  features {}
}

resource "azurerm_application_insights" "test" {
  name                = "mock"
  location            = "uksouth"
  resource_group_name = "mock"
  application_type    = "web"
}
