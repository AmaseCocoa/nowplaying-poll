FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

RUN mkdir /data && chown 65532:65532 /data

FROM gcr.io/distroless/static-debian13:nonroot AS runner
WORKDIR /

COPY --from=builder /app/main .
COPY --from=builder --chown=nonroot:nonroot /data /data

USER nonroot:nonroot
ENTRYPOINT ["./main"]
