#!/bin/bash -eux

sed dp-dd-job-creator-api-stub/appspec.yml dp-dd-job-creator-api-stub/scripts/codedeploy/* -i \
  -e s/\${CODEDEPLOY_USER}/$CODEDEPLOY_USER/g                         \
  -e s/^CONFIG_BUCKET=.*/CONFIG_BUCKET=$CONFIGURATION_BUCKET/         \
  -e s/^ECR_REPOSITORY_URI=.*/ECR_REPOSITORY_URI=$ECR_REPOSITORY_URI/ \
  -e s/^GIT_COMMIT=.*/GIT_COMMIT=$(cat build/revision)/               \
  -e s/^AWS_REGION=.*/AWS_REGION=$AWS_REGION/

mkdir -p artifacts/scripts/codedeploy

cp dp-dd-job-creator-api-stub/appspec.yml artifacts/
cp dp-dd-job-creator-api-stub/scripts/codedeploy/* artifacts/scripts/codedeploy
