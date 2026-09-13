FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
COPY --from=builder /out/server /server
EXPOSE 8084
ENTRYPOINT ["/server"]
