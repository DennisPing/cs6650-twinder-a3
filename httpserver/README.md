# HTTP Server

Load balanced achieved via Railway's horizontal scaling

## Build

```bash
docker build -t mushufeels/httpserver .
```

## Environment Variables

Make sure to add the environment variables either in an .env file or into your cloud platform console

1. AXIOM_API_TOKEN (api key)
2. AXIOM_DATASET (dataset name)
3. LOG_LEVEL (debug, info, warn, etc)
4. RABBITMQ_HOST (ip address)
5. AWS_ACCESS_KEY_ID (aws id)
6. AWS_SECRET_ACCESS_KEY (aws access key)

AXIOM_API_TOKEN=xaat-699934db-ee88-473a-b4e2-6d3a68a7c373
AXIOM_DATASET=a1-server-metrics
LOG_LEVEL=warn
RABBITMQ_HOST=3.132.190.142
AWS_ACCESS_KEY_ID=AKIAYZ7J7XHKHZIPMHSC
AWS_SECRET_ACCESS_KEY=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

## Generate Mocks

```bash
go generate ./...
```

## Run Tests

```bash
go test ./...
```