FROM golang:1.20.3-alpine3.16 AS builder
RUN apk --no-cache add build-base
ARG TARGETARCH
WORKDIR /src
RUN go install golang.org/x/vuln/cmd/govulncheck@latest
COPY src/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -a -ldflags="-s -w" -installsuffix cgo -v -o /chat-hvac .

FROM builder AS test
RUN go test ./...
RUN govulncheck ./...

FROM scratch AS final
COPY --from=builder /chat-hvac /chat-hvac
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/chat-hvac"]