FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /notification-api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget
COPY --from=build /notification-api /notification-api
EXPOSE 8084
USER 65532:65532
ENTRYPOINT ["/notification-api"]
