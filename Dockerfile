FROM golang:1.21.5-alpine AS builder
WORKDIR /app
COPY cmd/ ./cmd/
COPY internal/ ./internal
COPY go.mod .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main ./cmd/server/

FROM scratch AS runner
WORKDIR /
COPY --from=builder /app/main /main

ENTRYPOINT ["./main"]