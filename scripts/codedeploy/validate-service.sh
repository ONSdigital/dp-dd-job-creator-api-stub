#!/bin/bash

if [[ $(docker inspect --format="{{ .State.Running }}" dp-dd-job-creator-api-stub) == "false" ]]; then
  exit 1;
fi
