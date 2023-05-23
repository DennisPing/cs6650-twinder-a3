# AWS Linux 2 on Arm64

## If using an AWS Graviton Instance (arm64)

You cannot build a Docker image on an amd64 computer and use it on arm64 instance. Either use `buildx` to do cross platform build, or just clone the repo and build it natively.

### Setup dependencies
```bash
sudo yum update -y
sudo yum install -y docker
sudo usermod -a -G docker ec2-user
newgrp docker
# Log out and log back in here to ensure this takes effect
sudo systemctl enable docker
sudo systemctl start docker
```

### Build Docker image natively
```bash
git clone https://github.com/DennisPing/cs6650-twinder-a2.git
cd cs6650-twinder-a2
cd consumer
docker build -t mushufeels/consumer .
docker run -d --name consumer --env-file ~/consumer.env -p 8080:8080 mushufeels/consumer
```