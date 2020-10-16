#!/usr/bin/env bash

set -e

readonly _PROTOC_VERSION=3.13.0
readonly _DOCKER_IMAGE=go-proto-builder

readonly _GO_PB_PATH=github.com/golang/protobuf
readonly _VALIDATE_PATH=github.com/envoyproxy/protoc-gen-validate

function main {
    if [ -z "${TRANSSHIPMENT_ROOT}" ]; then
        echo "TRANSSHIPMENT_ROOT must be defined" >&2
        exit 1
    fi

    pushd "${TRANSSHIPMENT_ROOT}"
    trap popd EXIT

    _build_docker

    find pb -name '*pb*.go' -delete

    for proto in $(find pb -name '*.proto'); do
       echo "generating code from ${proto}" >&2
       docker run --rm \
            -v "${PWD}/pb/:/defs/" \
            "${_DOCKER_IMAGE}" \
            -I/defs/ \
            --go_out="paths=source_relative,plugins=grpc:/defs/" \
            --validate_out="lang=go,paths=source_relative:/defs/" \
            $(sed 's!^pb/!/defs/!' <<< "${proto}")

    done
}

function _build_docker {
    if [[ -z "${GOLANG_PROTOBUF_VERSION}" ]]; then
        local pb_version=$(go list -f '{{ .Version }}' -m "${_GO_PB_PATH}")
    else
        local pb_version="${GOLANG_PROTOBUF_VERSION}"
        echo "override set for ${_GO_PB_PATH}: ${pb_version}" >&2
    fi

    if [[ -z "${ENVOYPROXY_PROTOC_GEN_VALIDATE_VERSION}" ]]; then
        local validate_version=$(go list -f '{{ .Version }}' -m "${_VALIDATE_PATH}")
    else
        local validate_version="${ENVOYPROXY_PROTOC_GEN_VALIDATE_VERSION}"
        echo "override set for ${_VALIDATE_PATH}: ${validate_version}" >&2
    fi

    docker build -t "${_DOCKER_IMAGE}" -f - . << EOF
FROM golang:latest

RUN apt-get update \
    && apt-get install unzip \
    && wget \
        -O /tmp/protoc.zip \
        https://github.com/protocolbuffers/protobuf/releases/download/v${_PROTOC_VERSION}/protoc-${_PROTOC_VERSION}-linux-x86_64.zip \
    && mkdir -p /tmp/protoc /usr/local/include \
    && unzip -d /tmp/protoc /tmp/protoc.zip \
    && mv /tmp/protoc/bin/protoc /usr/local/bin/protoc \
    && mv /tmp/protoc/include/google /usr/local/include/google \
    && go get -d ${_GO_PB_PATH}/protoc-gen-go \
    && git -C /go/src/${_GO_PB_PATH} checkout ${pb_version} \
    && (cd /go/src/${_GO_PB_PATH}/protoc-gen-go && go install) \
    && go get -d ${_VALIDATE_PATH} \
    && git -C /go/src/${_VALIDATE_PATH} checkout ${validate_version} \
    && make -C /go/src/${_VALIDATE_PATH} build

ENTRYPOINT [ \
    "protoc", \
     "-I/usr/local/include", \
     "-I/go/src/${_VALIDATE_PATH}" \
]
EOF
}

main