FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /order-api ./cmd/api
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /order-api /order-api
USER 65532:65532
EXPOSE 8083
ENTRYPOINT ["/order-api"]
