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

FROM node:4-alpine

WORKDIR /function

# Install ImageMagick and AWS SDK as provided by Lambda.
RUN apk update && apk --no-cache add imagemagick
RUN npm install aws-sdk@2.2.32 imagemagick && npm cache clear

# cli should forbid this name
ADD bootstrap.js /function/lambda-bootstrap.js

# Run the handler, with a payload in the future.
ENTRYPOINT ["node", "./lambda-bootstrap"]
