FROM golang:latest

WORKDIR /app

RUN apt update && apt install -y ffmpeg

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o server ./cmd/server/
CMD ["./server"]
