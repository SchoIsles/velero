# Copyright 2020 the Velero contributors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
FROM hub.qucheng.com/library/god as builder-env

ARG GOPROXY
ARG PKG
ARG VERSION
ARG GIT_SHA
ARG GIT_TREE_STATE
ARG REGISTRY

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=${GOPROXY} \
    LDFLAGS="-X ${PKG}/pkg/buildinfo.Version=${VERSION} -X ${PKG}/pkg/buildinfo.GitSHA=${GIT_SHA} -X ${PKG}/pkg/buildinfo.GitTreeState=${GIT_TREE_STATE} -X ${PKG}/pkg/buildinfo.ImageRegistry=${REGISTRY}"

WORKDIR /go/src/github.com/vmware-tanzu/velero

COPY . /go/src/github.com/vmware-tanzu/velero

RUN sed -i -r 's/(deb|security).debian.org/mirrors.tencent.com/g' /etc/apt/sources.list && apt-get update && apt-get install -y bzip2

FROM --platform=$BUILDPLATFORM builder-env as builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG PKG
ARG BIN
ARG RESTIC_VERSION

ENV GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT}

RUN export GOARM=$( echo "${GOARM}" | cut -c2-) && \
    go build -o /output/${BIN} \
    -ldflags "${LDFLAGS}" ${PKG}/cmd/${BIN}

FROM hub.qucheng.com/library/debian:11.3-slim

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

RUN sed -i -r 's/(deb|security).debian.org/mirrors.cloud.tencent.com/g' /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y curl wget tzdata zip unzip s6 pwgen cron apt-transport-https ca-certificates procps \
    && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && dpkg-reconfigure --frontend noninteractive tzdata \
    && apt-get clean

COPY --from=builder /output/velero /velero
COPY --from=hub.qucheng.com/third-party/restic:0.13.1 /usr/bin/restic /usr/bin/restic

