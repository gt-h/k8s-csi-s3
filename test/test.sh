#!/bin/sh
export MINIO_ACCESS_KEY=FJDSJ
export MINIO_SECRET_KEY=DSG643HGDS

mkdir -p /tmp/minio
minio server /tmp/minio &>/dev/null &
sleep 5
#go test ./... -cover -ginkgo.noisySkippings=false -ginkgo.skip="should fail when requesting to create a volume with already existing name and different capacity"

# Run ordinary tests in all packages, excluding the Ginkgo suite.
go test ./... -cover -count=1 -skip '^TestS3Driver$' || exit 1

# Run the Ginkgo suite with its package-specific flags.
go test ./pkg/driver -cover -count=1 -run '^TestS3Driver$' \
  -ginkgo.noisySkippings=false \
  -ginkgo.skip="should fail when requesting to create a volume with already existing name and different capacity"
