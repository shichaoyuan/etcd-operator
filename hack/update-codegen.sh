#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

bash ../vendor/k8s.io/code-generator/generate-groups.sh \
  "all" \
  etcd-operator/pkg/generated \
  etcd-operator/pkg/apis \
  samplecrd:v1 \
  --go-header-file $(pwd)/boilerplate.go.txt \
  --output-base $(pwd)/../../