ARG BUILDER_IMAGE=golang:1.24.2-alpine
ARG IMAGE=debian:bookworm-slim

FROM ${BUILDER_IMAGE} as builder

WORKDIR $GOPATH/src/mypackage/myapp/

COPY ./go.mod .

RUN apk update && apk add git gcc build-base gcompat

ENV GO111MODULE=on

ENV LD_LIBRARY_PATH=/root/.cache/chroma/shared/libtokenizers:/root/.cache/chroma/shared/onnxruntime:$LD_LIBRARY_PATH

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
        -ldflags='-w -s -extldflags "-static"' -a \
        -o /go/bin/sashimi-backend 

FROM ${IMAGE}

COPY --from=builder /go/bin/sashimi-backend /go/bin/sashimi-backend

RUN apt-get update
RUN apt-get install ca-certificates -y && update-ca-certificates

ENTRYPOINT ["/go/bin/sashimi-backend", "serve"]
