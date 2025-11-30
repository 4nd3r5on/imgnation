================
==== CONFIG ====
================

CONFIG_PATH   Optional. Default "./config.yml"

# Auth Configuration Environment Variables

AUTH_JWT_SECRET   Required. JWT secret

# API Configuration Environment Variables

API_HOST     Optional. Default 127.0.0.1
API_PORT     Optional. Default 80
API_PREFIX   Optional. Default ""

# Minio Configuration Environment Variables

MINIO_CREDENTIALS_FILE_PATH   Optional. Path to credentials file. If set, file-based auth is used.
ACCESS_KEY                    Optional. Minio access key (used if no credentials file is provided).
SECRET_KEY                    Optional. Minio secret key (used if no credentials file is provided).
MINIO_ENDPOINT                Required. Minio server endpoint (e.g., "s3.example.com:9000").
MINIO_USE_SSL                 Optional. Whether to use SSL (true/false). Defaults to false.

# Mongo Configuration Environment Variables

MONGO_HOST       Optional. MongoDB host. Defaults to value of defaultMongoHost.
MONGO_PORT       Optional. MongoDB port. Defaults to value of defaultMongoPort.
MONGO_USER       Optional. MongoDB username. Defaults to value of defaultMongoUser.
MONGO_PASSWORD   Required. MongoDB password.
MONGO_DB         Optional. MongoDB database name. Defaults to value of defaultMongoDB.




NOTES:

ffmpeg -y -i /tmp/uploads/<upload_id> \
  -filter_complex \
    "[0:v]split=3[v1080][v720][v480]; \
     [v1080]scale='min(1920,iw)':-1[v1080out]; \
     [v720] scale='min(1280,iw)':-1[v720out]; \
     [v480] scale='min(854,iw)':-1[v480out]" \
  -map "[v1080out]" -c:v:0 libx264 -b:v:0 5000k -maxrate:v:0 5350k -bufsize:v:0 7500k -preset medium -g 50 -keyint_min 50 -sc_threshold 0 \
  -map "[v720out]"  -c:v:1 libx264 -b:v:1 2800k -maxrate:v:1 3000k -bufsize:v:1 4200k -preset medium -g 50 -keyint_min 50 -sc_threshold 0 \
  -map "[v480out]"  -c:v:2 libx264 -b:v:2 1200k -maxrate:v:2 1280k -bufsize:v:2 1800k -preset medium -g 50 -keyint_min 50 -sc_threshold 0 \
  -map a:0 -c:a:0 aac -b:a:0 128k \
  -f dash -seg_duration 5 -use_template 1 -use_timeline 1 \
  -init_seg_name 'init-stream$RepresentationID$.m4s' \
  -media_seg_name 'chunk-stream$RepresentationID$-$Number%05d$.m4s' \
  -adaptation_sets "id=0,streams=v id=1,streams=a" \
  /tmp/video_processing/<upload_id>/dummy.mpd
