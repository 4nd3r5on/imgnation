source .env

docker compose -f ./docker-compose.yml build

docker compose -f ./docker-compose.yml up minio

Host
docker compose -f ./docker-compose.yml exec minio "/bin/bash"

Container

mc alias set prod http://127.0.0.1:9000 \
  "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc admin accesskey create prod/

Added `dev` successfully.
Access Key: xxx
Secret Key: xxx
Expiration: NONE
Name:

And then or add to app in compose as env variable
ACCESS_KEY=xxx
SECRET_KEY=xxx

Or add ./imgnation-backend/credentials.json
{
  "accessKey": "xxx",
  "secretKey": "xxx"
}

docker compose -f ./docker-compose.yml up
