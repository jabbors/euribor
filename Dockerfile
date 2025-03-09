# defaults which may be overridden from the build command
ARG GO_VERSION=1.22
ARG ALPINE_VERSION=3.19

# build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

COPY gorates /go/src/github.com/jabbors/gorates
WORKDIR /go/src/github.com/jabbors/gorates
ARG APP_VERSION=0.0
RUN go install -ldflags="-X \"main.version=${APP_VERSION}\""

# final stage
FROM alpine:${ALPINE_VERSION}

RUN apk --no-cache add ca-certificates bash curl
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=builder /go/bin/gorates /usr/bin/gorates
COPY gorates/chart-icon* /usr/bin/
COPY fetch-parse-insert.sh /usr/bin/
USER nobody:nobody
CMD [ "/usr/bin/gorates" ]
EXPOSE 8080
