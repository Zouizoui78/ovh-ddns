FROM --platform=$BUILDPLATFORM golang:1.26.3-alpine AS base

WORKDIR /build
COPY go.mod go.sum* ./
RUN apk add --no-cache make && \
    go mod download
COPY . .

FROM base AS build
RUN make

FROM base AS test
RUN make test-cicd

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

COPY --from=build /build/bin/ovh-ddns /ovh-ddns
WORKDIR /data
VOLUME ["/data"]

ENTRYPOINT ["/ovh-ddns"]
