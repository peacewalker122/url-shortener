FROM golang:1.26 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/server ./cmd/server

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/bin/server /app/server

EXPOSE 8000

CMD ["/app/server"]
