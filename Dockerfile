FROM golang:1.22-alpine3.19 as builder
WORKDIR /build
ARG version
ENV version_env=$version
ARG app_name
ENV app_name_env=$app_name
COPY . .
RUN apk update && apk upgrade && apk add --no-cache sqlite-dev
RUN go build -ldflags="-X 'main.version=$version_env'" -o /main .

FROM alpine:3.19
WORKDIR /app
ARG app_name
ENV app_name_env=$app_name
COPY --from=builder main /app/$app_name_env
COPY /conf/config.yml /app/config.yml
COPY /conf/default_remote_config.json* /app/default_remote_config.json
COPY /migrations /app/migrations
ENTRYPOINT /app/$app_name_env
