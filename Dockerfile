FROM golang:1.25.3 AS build 

WORKDIR /shortener
COPY . .

ENV CGO_ENABLED=0

RUN go build -o url-shortener ./cmd/server/

FROM alpine:latest AS app

WORKDIR /shortener

COPY --from=build ./shortener/url-shortener ./
COPY --from=build ./shortener/static ./static
COPY --from=build ./shortener/templates ./templates

EXPOSE 8080/tcp

ENTRYPOINT [ "./url-shortener" ]