FROM golang:1.18.3-alpine3.16 as builder
WORKDIR /copymediagroup
COPY . .
RUN go build -ldflags="-w -s" .
RUN rm -rf *.go && rm -rf go.*
FROM alpine:3.16.0
COPY --from=builder /copymediagroup/copymediagroup /copymediagroup
ENTRYPOINT ["/copymediagroup"]
