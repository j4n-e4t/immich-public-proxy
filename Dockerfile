FROM docker.io/golang:1.24.2-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main .

# Start a new stage using alpine for ca-certificates
FROM alpine/git:latest AS certs

FROM scratch

# Copy the ca-certificates from the certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/main /app/main

ENV IMMICH_BASE_URL=http://immich:8000/
ENV TS_HOSTNAME=immich-public-proxy
ENV TS_AUTHKEY=your-key

CMD [ "/app/main" ]