FROM alpine:latest AS git-stage
RUN apk add git

ARG EXTCOMMIT
ARG EXTVERSION

WORKDIR /wirejump
COPY . .

# check if build args have been provided; use local git info otherwise
RUN if [[ -n "$EXTCOMMIT" ]]; then \
    echo ${EXTCOMMIT} > /commit-file; \
else \
    echo $(git rev-parse --short HEAD 2>/dev/null || echo "") > /commit-file; \
fi
RUN if [[ -n "$EXTVERSION" ]]; then \
    echo ${EXTVERSION} > /version-file; \
else \
    echo $(git describe --tags 2>/dev/null || echo "") > /version-file; \
fi

# code requires 1.18 or later
FROM golang:latest AS build-stage

WORKDIR /wirejump

COPY --from=git-stage /commit-file  /
COPY --from=git-stage /version-file /

COPY wirejump .
RUN COMMIT="$(cat /commit-file)" VERSION="$(cat /version-file)" OUTPUT_DIR=/build GOOS=linux make
RUN ls /build/

FROM scratch AS export-stage
COPY --from=build-stage /build /
