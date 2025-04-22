FROM golang:1.24.1-alpine AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o swift-api ./cmd

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/swift-api /app/swift-api
COPY --from=build /app/assets /app/assets

EXPOSE 8080
CMD ["/app/swift-api"]
