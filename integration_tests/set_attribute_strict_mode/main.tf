terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "= 2.70"
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

resource "azurerm_storage_account" "test" {
  name                     = "mockstorageaccount"
  resource_group_name      = "mock"
  location                 = "uksouth"
  account_tier             = "Standard"
  account_replication_type = "LRS"
}
