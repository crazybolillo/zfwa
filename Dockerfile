FROM golang:1.23 AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /bin/zfwa

FROM scratch

COPY --from=builder /bin/zfwa /bin/zfwa

ENTRYPOINT ["zfwa"]
