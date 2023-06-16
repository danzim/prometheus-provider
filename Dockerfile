ARG BUILDPLATFORM="linux/amd64"
ARG BUILDERIMAGE="golang:1.19-bullseye"
ARG BASEIMAGE="gcr.io/distroless/static:nonroot"

FROM --platform=${BUILDPLATFORM} ${BUILDERIMAGE} as builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""
ARG LDFLAGS

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

WORKDIR /go/src/github.com/danzim/prometheus-provider

COPY . .

RUN make build

FROM ${BASEIMAGE}

WORKDIR /

COPY --from=builder /go/src/github.com/danzim/prometheus-provider/bin/provider .

USER 65532:65532

ENTRYPOINT ["/provider"]