FROM golang:1.25-alpine AS dev
WORKDIR /app
# gcc/musl-dev は `go test -race`（cgo 必須）のために入れる
RUN apk add --no-cache gcc musl-dev
# air は v1.67.2 以降 Go >= 1.26 を要求するため、Go 1.25 対応の最終版に固定する
RUN go install github.com/air-verse/air@v1.67.1
COPY go.mod go.sum ./
RUN go mod download
EXPOSE 8080
CMD ["air"]

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/server

FROM gcr.io/distroless/static-debian12 AS prod
COPY --from=builder /app/bin/server /server
EXPOSE 8080
CMD ["/server"]
