# [ADKとRAG Engineのハンズオン](https://zenn.dev/soundtricker/articles/a983714ac4e04a)

## Run locally

```shell
uv run adk web --reload
```

## Deploy to Agent Engine

```shell
uv run adk deploy agent_engine \
  --project=PROJECT_ID \
  --region=us-west1 \
  --staging_bucket=gs://PROJECT_ID-staging \
  --display_name="My Agent" \
  ./concierge
```
