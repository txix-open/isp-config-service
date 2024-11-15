FROM golang:1.23-alpine3.20 as builder
WORKDIR /build
ARG version
ENV version_env=$version
ARG app_name
ENV app_name_env=$app_name
COPY . .
RUN apk update && apk upgrade && apk add --no-cache gcc musl-dev
RUN go build -ldflags="-X 'main.version=$version_env'" -o /main .

FROM alpine:3.20

RUN apk add --no-cache tzdata
RUN cp /usr/share/zoneinfo/Europe/Moscow /etc/localtime
RUN echo "Europe/Moscow" > /etc/timezone

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

WORKDIR /app
ARG app_name
ENV app_name_env=$app_name
COPY --from=builder main /app/$app_name_env
COPY /conf/config.yml /app/config.yml
COPY /conf/default_remote_config.json* /app/default_remote_config.json
COPY /migrations /app/migrations
ENTRYPOINT /app/$app_name_env
