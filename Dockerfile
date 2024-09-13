FROM golang:1.23-alpine AS builder

RUN apk update; \
    apk upgrade; \
    apk add --no-cache ca-certificates; \
    update-ca-certificates

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /bin/zfwa

FROM scratch

COPY --from=builder /bin/zfwa /bin/zfwa
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["zfwa"]
