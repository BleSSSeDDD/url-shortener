FROM golang:1.25.3 AS build 

WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go build -o url-shortener ./cmd/server/

FROM alpine:latest AS app

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=build /app/url-shortener ./
COPY --from=build /app/static ./static
COPY --from=build /app/templates ./templates

USER app

EXPOSE 8080
ENTRYPOINT [ "./url-shortener" ]