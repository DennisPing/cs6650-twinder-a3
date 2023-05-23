# RabbitMQ Consumer

## Build and Push up to Docker Hub
```bash
docker build -t mushufeels/consumer .
docker push mushufeels/consumer:latest
```

## Pull down from Docker Hub into VM
```bash
docker pull mushufeels/consumer:latest
```

## Create .env file in VM
```bash
touch ~/consumer.env
echo "RABBITMQ_HOST={ip_address}" >> ~/consumer.env
echo "LOG_LEVEL=warn" >> ~/consumer.env
```

## Run container
```bash
docker run -d --name consumer --env-file ~/consumer.env -p 8080:8080 mushufeels/consumer
```

## Stop containers
```bash
docker stop {container_name}
```

## Restart stopped containers
```bash
docker start {container_name}
```

## View logs of containers
```bash
docker logs {container_name}
```

## Remove containers
```bash
docker rm {container_name}
```

## Remove images
```bash
docker image rm {image_id}
```