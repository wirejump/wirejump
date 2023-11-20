FROM alpine:latest AS git-stage
RUN apk add git

WORKDIR /wirejump
COPY . .

# get version info from git
RUN echo $(git rev-parse --short HEAD 2>/dev/null || echo "") > /commit-file
RUN echo $(git describe --tags 2>/dev/null || echo "") > /version-file

# code requires 1.18 or later
FROM golang:latest AS build-stage

WORKDIR /wirejump

COPY --from=git-stage /commit-file  /
COPY --from=git-stage /version-file /

RUN cat /commit-file
RUN cat /version-file

COPY wirejump .
RUN COMMIT="$(cat /commit-file)" VERSION="$(cat /version-file)" OUTPUT_DIR=/build GOOS=linux make
RUN ls /build/

FROM scratch AS export-stage
COPY --from=build-stage /build /
