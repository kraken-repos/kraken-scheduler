#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export REPO_ROOT_DIR=$(dirname $0)/..

CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${REPO_ROOT_DIR}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

KNATIVE_CODEGEN_PKG=${KNATIVE_CODEGEN_PKG:-$(cd ${REPO_ROOT_DIR}; ls -d -1 . 2>/dev/null || echo ../pkg)}

# Scheduler
API_DIRS_SOURCES=(pkg)

for DIR in "${API_DIRS_SOURCES[@]}"; do
  # generate the code with:
  # --output-base    because this script should also be able to run inside the vendor dir of
  #                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
  #                  instead of the $GOPATH directly. For normal projects this can be dropped.
  ${CODEGEN_PKG}/generate-groups.sh "deepcopy,client,informer,lister" \
    "kraken.dev/kraken-scheduler/${DIR}/client" "kraken.dev/kraken-scheduler/${DIR}/apis" \
    "scheduler:v1alpha1" \
    --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt

  # Knative Injection
  ${KNATIVE_CODEGEN_PKG}/hack/generate-knative.sh "injection" \
    "kraken.dev/kraken-scheduler/${DIR}/client" "kraken.dev/kraken-scheduler/${DIR}/apis" \
    "scheduler:v1alpha1" \
    --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt
done

# Depends on generate-groups.sh to install bin/deepcopy-gen
${GOPATH}/bin/deepcopy-gen \
  -O zz_generated.deepcopy \
  --go-header-file ${REPO_ROOT_DIR}/hack/boilerplate.go.txt \
  -i kraken.dev/kraken-scheduler/pkg/apis \

# Make sure our dependencies are up-to-date
#${REPO_ROOT_DIR}/hack/update-deps.sh
