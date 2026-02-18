#!/bin/bash

go mod vendor
if ! git ls-files --error-unmatch vendor/modules.txt >/dev/null 2>&1; then
  echo "Vendoring is enabled but vendor/modules.txt is not tracked. Commit vendor/ (or at least vendor/modules.txt)."
  exit 1
fi
if ! git diff --exit-code -- vendor/modules.txt; then
  echo "Vendoring is out of sync. Run 'go mod vendor' (or 'make vendor') and commit the changes."
  git diff -- vendor/modules.txt | head -n 200
  exit 1
fi

go vet ./...

go test ./...

go test -race ./...
