#!/usr/bin/env bash
# Deletes the local kind cluster entirely, freeing all resources.
set -euo pipefail

kind delete cluster --name observability
