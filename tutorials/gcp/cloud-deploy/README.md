# Cloud Deploy Tutorial for Cloud Run

[cloud-deploy-tutorials/tutorials/e2e-run at main · GoogleCloudPlatform/cloud-deploy-tutorials](https://github.com/GoogleCloudPlatform/cloud-deploy-tutorials/tree/main/tutorials/e2e-run)

```shell
export GOOGLE_CLOUD_PROJECT="your-project-id"
gcloud config set project $GOOGLE_CLOUD_PROJECT
gcloud services enable storage.googleapis.com compute.googleapis.com artifactregistry.googleapis.com
```