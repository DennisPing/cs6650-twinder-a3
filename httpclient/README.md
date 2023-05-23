# HTTP Client

Single client using 50 goroutines and sending 500k total POST requests

## Build

```bash
docker build -t mushufeels/httpclient .
```

## Environment Variables

Make sure to add the environment variables either in a .env file or into your PaaS console

1. LOG_LEVEL (debug, info, warn, etc)
2. SERVER_URL (https://server_url or http://server_ip:port)
