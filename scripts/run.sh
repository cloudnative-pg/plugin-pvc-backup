#!/usr/bin/env bash
##
## Copyright The CloudNativePG Contributors
##
## Licensed under the Apache License, Version 2.0 (the "License");
## you may not use this file except in compliance with the License.
## You may obtain a copy of the License at
##
##     http://www.apache.org/licenses/LICENSE-2.0
##
## Unless required by applicable law or agreed to in writing, software
## distributed under the License is distributed on an "AS IS" BASIS,
## WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
## See the License for the specific language governing permissions and
## limitations under the License.
##

set -eu

cd "$(dirname "$0")/.." || exit

if [ -f .env ]; then
    source .env
fi

# The following script deployes this plugin in a locally running CNPG installation
# in Kind
current_context=$(kubectl config view --raw -o json | jq -r '."current-context"' | sed "s/kind-//")
kind load docker-image --name=${current_context} plugin-pvc-backup:${VERSION:-latest}

kubectl patch deployment -n cnpg-system  cnpg-controller-manager --patch-file  kubernetes/deployment-patch.json
kubectl rollout restart deployment -n cnpg-system  cnpg-controller-manager
kubectl rollout status deployment -n cnpg-system  cnpg-controller-manager
