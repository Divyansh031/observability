FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server  
# Disable Cgo. Produces a fully static binary that does not need any C libraries. Perfect for a tiny Alpine image.

FROM alpine:3.20
COPY --from=builder /out/server /server
EXPOSE 8082
ENTRYPOINT ["/server"]
