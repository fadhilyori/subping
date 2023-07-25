FROM golang:1.20-alpine AS builder

ARG VERSION=dev

# hadolint ignore=DL3018
RUN apk add --no-cache make

WORKDIR /app

COPY --link go.mod /app/go.mod
COPY --link go.sum /app/go.sum

RUN go mod download

COPY --link . .

RUN make build VERSION=${VERSION}

FROM scratch

COPY --link --from=builder /app/out/subping /subping

ENTRYPOINT [ "/subping" ]
