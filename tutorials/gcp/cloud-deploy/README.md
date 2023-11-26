# Cloud Deploy Tutorial for Cloud Run

[cloud-deploy-tutorials/tutorials/e2e-run at main · GoogleCloudPlatform/cloud-deploy-tutorials](https://github.com/GoogleCloudPlatform/cloud-deploy-tutorials/tree/main/tutorials/e2e-run)

```shell
export GOOGLE_CLOUD_PROJECT="your-project-id"
gcloud config set project $GOOGLE_CLOUD_PROJECT
gcloud auth application-default login
skaffold build --interactive=false --default-repo us-central1-docker.pkg.dev/${GOOGLE_CLOUD_PROJECT}/hello-app --file-output artifacts.json
gcloud deploy releases create "$(date +%Y%m%d%H%M%S)" \
  --region us-central1 \
  --delivery-pipeline hello-app \
  --build-artifacts artifacts.json \
  --gcs-source-staging-dir gs://us-central1.deploy-artifacts.$GOOGLE_CLOUD_PROJECT.appspot.com/source \
  --source .
```