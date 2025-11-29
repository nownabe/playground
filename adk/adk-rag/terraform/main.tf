terraform {
  required_version = "1.14.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.12.0"
    }
  }
}

variable "project" {
  type = string
}

provider "google" {
  project = var.project
  region  = "us-west1"
}

locals {
  services = [
    "aiplatform",
    "cloudbuild",
    "run",
  ]
}

resource "google_project_service" "services" {
  for_each = toset(local.services)

  service = "${each.value}.googleapis.com"
}

resource "google_vertex_ai_rag_engine_config" "default" {
  rag_managed_db_config {
    basic {}
  }
}
