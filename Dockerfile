#
# Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
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
#

# build stage
FROM golang:1.16.3-alpine3.12 AS build-env
RUN apk add --no-cache gcc musl-dev
ARG D=/go/src/github.com/fnproject/cli
ARG GO111MODULE=on
ARG GOFLAGS=-mod=vendor
ADD . $D
RUN cd $D && go build -o fn-alpine && cp fn-alpine /tmp/

# final stage
FROM alpine:3.12
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=build-env /tmp/fn-alpine /app/fn
ENTRYPOINT ["./fn"]
