#!/bin/bash

# Copyright 2017 The Kubernetes Authors.
# Copyright 2018 Chef Software Inc. and/or applicable contributors.
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

# Based on https://github.com/kubernetes/code-generator/blob/893b4433a4ba929dd0dacebf7c8956682d7a5d5f/hack/update-codegen.sh

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
readonly CODEGEN_DIR="${SCRIPT_ROOT}/vendor/k8s.io/code-generator"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
"${CODEGEN_DIR}"/generate-groups.sh all \
  github.com/biome-sh/biome-operator/pkg/client github.com/biome-sh/biome-operator/pkg/apis \
  biome:v1beta1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt
