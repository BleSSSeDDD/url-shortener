FROM golang:1.25.3 AS build

WORKDIR /build

# сначала только зависимости
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -o url-shortener ./cmd/server/

FROM alpine:latest AS server

RUN addgroup -S app && adduser -S app -G app

WORKDIR /server

COPY --from=build /build/static ./static
COPY --from=build /build/templates ./templates 
COPY --from=build /build/url-shortener ./url-shortener

USER app

EXPOSE 8080

ENTRYPOINT [ "./url-shortener"]