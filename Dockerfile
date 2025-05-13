ARG BUILDER_IMAGE=golang:1.24.2-alpine
ARG DISTROLESS_IMAGE=gcr.io/distroless/base-debian12

FROM ${BUILDER_IMAGE} as builder

WORKDIR $GOPATH/src/mypackage/myapp/

COPY ./go.mod .

RUN apk update && apk add git gcc build-base ca-certificates gcompat

ENV GO111MODULE=on
RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' -a \
        -o /go/bin/sashimi-backend 

FROM ${DISTROLESS_IMAGE}

COPY --from=builder /go/bin/sashimi-backend /go/bin/sashimi-backend

ENTRYPOINT ["/go/bin/sashimi-backend", "serve"]
