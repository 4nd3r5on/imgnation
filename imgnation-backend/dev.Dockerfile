FROM golang:latest

WORKDIR /app

# Install ffmpeg
RUN apt update && apt install -y ffmpeg

# Install Air for hot-reloading
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

CMD ["air"]
