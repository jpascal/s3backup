FROM golang:1.24.1-alpine AS builder

WORKDIR /builder

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./ /builder

RUN go build -trimpath -o ./dist/bin/s3backup cmd/s3backup/main.go

FROM alpine:3.20.3

WORKDIR /app

COPY --from=builder /builder/dist /app/

ENTRYPOINT ["/app/bin/s3backup"]

