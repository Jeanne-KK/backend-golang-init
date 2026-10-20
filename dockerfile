# Dev image: bind-mount โค้ดที่ /app แล้วรัน hot reload ด้วย Air
# Build:  docker build -f dockerfile -t initbackgo-dev .
# Run:    docker run --rm -p 8080:8080 --env-file .env -v "$(pwd):/app" initbackgo-dev
# ใน container ตั้ง DB_HOST เป็น host ของ MySQL (เช่น host.docker.internal หรือชื่อ service ใน compose)

FROM golang:1.26-bookworm

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
