FROM golang:1.13-alpine3.10 AS base_image

FROM base_image AS build

RUN apk add --no-cache ca-certificates curl git build-base
RUN mkdir drm-proxy
RUN curl -sL https://github.com/cbsinteractive/drm-proxy/tarball/master | tar -C /drm-proxy --strip 1 -xz
RUN cd drm-proxy && go build

FROM base_image

RUN apk add --no-cache ca-certificates
RUN mkdir /usr/local/drm-proxy/

COPY --from=build drm-proxy /usr/local/drm-proxy/

ENV GCS_HELPER_BUCKET_NAME=tsymborski-testing
ENV GCS_HELPER_PROXY_PREFIX=/proxy
ENV GCS_HELPER_MAP_PREFIX=/map
ENV GCS_HELPER_MAP_REGEX_FILTER="\.(mp4|srt|vtt)$"
ENV GCS_HELPER_LOG_LEVEL=debug
ENV GOOGLE_APPLICATION_CREDENTIALS=/etc/google-creds.json

ENTRYPOINT ["/usr/local/gcs-helper/gcs-helper"]