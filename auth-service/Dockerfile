FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /auth-api ./cmd/api
FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /auth-api /auth-api
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/auth-api"]
