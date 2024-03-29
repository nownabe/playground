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
    "artifactregistry.googleapis.com",
    "cloudbuild.googleapis.com",
    "clouddeploy.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "compute.googleapis.com",
    "iam.googleapis.com",
    "run.googleapis.com",
    "storage.googleapis.com",
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

resource "google_service_account_iam_member" "hello-app-serviceAccountUser" {
  for_each = local.envs_map

  service_account_id = google_service_account.hello-app[each.key].name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.hello-app-deploy.email}"
}

resource "google_project_iam_member" "logWriter" {
  for_each = local.envs_map

  project = data.google_project.project.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.hello-app[each.key].email}"
}

resource "google_cloud_run_v2_service" "hello-app" {
  for_each = local.envs_map

  location = var.region
  name     = "hello-app-${each.key}"
  ingress  = "INGRESS_TRAFFIC_ALL"

  template {
    containers {
      image = "us-docker.pkg.dev/cloudrun/container/hello"
    }
    service_account = google_service_account.hello-app[each.key].email
  }

  lifecycle {
    ignore_changes = [
      template
    ]
  }
}

resource "google_cloud_run_v2_service_iam_member" "hello-app-developer" {
  for_each = local.envs_map

  location = var.region
  name     = google_cloud_run_v2_service.hello-app[each.key].name
  role     = "roles/run.developer"
  member   = "serviceAccount:${google_service_account.hello-app-deploy.email}"
}

resource "google_cloud_run_v2_service_iam_member" "hello-app-invoker" {
  for_each = local.envs_map

  location = var.region
  name     = google_cloud_run_v2_service.hello-app[each.key].name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

// Cloud Deploy

resource "google_service_account" "hello-app-deploy" {
  account_id = "hello-app-deploy"
}

resource "google_project_iam_member" "deploy_logWriter" {
  project = data.google_project.project.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.hello-app-deploy.email}"
}

data "google_storage_bucket" "hello-app-artifact-storage" {
  name = "${var.region}.deploy-artifacts.${data.google_project.project.project_id}.appspot.com"
}

resource "google_storage_bucket_iam_member" "hello-app-artifact-storage" {
  bucket = data.google_storage_bucket.hello-app-artifact-storage.name
  role   = "roles/storage.objectUser"
  member = "serviceAccount:${google_service_account.hello-app-deploy.email}"
}

resource "google_clouddeploy_delivery_pipeline" "hello-app" {
  location    = var.region
  name        = "hello-app"
  description = "Delivery pipeline for hello-app"

  serial_pipeline {
    stages {
      target_id = google_clouddeploy_target.hello-app-dev.name
      profiles  = ["dev"]
    }

    stages {
      target_id = google_clouddeploy_target.hello-app-stg.name
      profiles  = ["stg"]
    }

    stages {
      target_id = google_clouddeploy_target.hello-app-prd.name
      profiles  = ["prd"]
    }
  }
}

locals {
  deploy_parameters = {
    dev = {
      service_name    = google_cloud_run_v2_service.hello-app["dev"].name
      service_account = google_service_account.hello-app["dev"].email
      name            = "DEV"
    }
    stg = {
      service_name    = google_cloud_run_v2_service.hello-app["stg"].name
      service_account = google_service_account.hello-app["stg"].email
      name            = "STG"
    }
    prd = {
      service_name    = google_cloud_run_v2_service.hello-app["prd"].name
      service_account = google_service_account.hello-app["prd"].email
      name            = "PRD"
    }
  }
}

resource "google_clouddeploy_target" "hello-app-dev" {
  location         = var.region
  name             = "hello-app-dev"
  description      = "dev environment for hello-app"
  require_approval = false
  run {
    location = "projects/${data.google_project.project.project_id}/locations/${var.region}"
  }
  execution_configs {
    usages          = ["RENDER", "DEPLOY"]
    service_account = google_service_account.hello-app-deploy.email
  }

  deploy_parameters = local.deploy_parameters["dev"]
}

resource "google_clouddeploy_target" "hello-app-stg" {
  location         = var.region
  name             = "hello-app-stg"
  description      = "stg environment for hello-app"
  require_approval = true
  run {
    location = "projects/${data.google_project.project.project_id}/locations/${var.region}"
  }
  execution_configs {
    usages          = ["RENDER", "DEPLOY"]
    service_account = google_service_account.hello-app-deploy.email
  }
  deploy_parameters = local.deploy_parameters["stg"]
}

resource "google_clouddeploy_target" "hello-app-prd" {
  location         = var.region
  name             = "hello-app-prd"
  description      = "prd environment for hello-app"
  require_approval = true
  run {
    location = "projects/${data.google_project.project.project_id}/locations/${var.region}"
  }
  execution_configs {
    usages          = ["RENDER", "DEPLOY"]
    service_account = google_service_account.hello-app-deploy.email
  }
  deploy_parameters = local.deploy_parameters["prd"]
}
