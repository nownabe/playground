terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.7.0"
    }
  }
}

provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
}

locals {
  enabled_services = [
    "cloudbuild.googleapis.com",
    "clouddeploy.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "iam.googleapis.com",
    "run.googleapis.com",
  ]

  enabled_services_map = {
    for s in local.enabled_services : s => true
  }
}

resource "google_project_service" "cloudbuild" {
  for_each = local.enabled_services_map

  project = data.google_project.project.project_id
  service = each.key
}

resource "google_artifact_registry_repository" "hello-app" {
  repository_id = "hello-app"
  location      = var.region
  format        = "DOCKER"
}

// Cloud Run

locals {
  envs = ["dev", "stg", "prd"]

  envs_map = {
    for e in local.envs : e => true
  }
}

resource "google_service_account" "hello-app" {
  for_each = local.envs_map

  account_id = "hello-app-${each.key}"
}

resource "google_project_iam_member" "logWriter" {
  for_each = local.envs_map

  project = data.google_project.project.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.hello-app[each.key].email}"
}

// Cloud Deploy

resource "google_service_account" "hello-app-deploy" {
  account_id = "hello-app-deploy"
}
