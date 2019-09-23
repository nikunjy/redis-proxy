FROM golang:latest
ARG EXPIRY
ENV CACHE_EXPIRY "$EXPIRY"
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o redis-proxy .
CMD ["./redis-proxy", "--proxy-port", "8081"]
