#!/usr/bin/env bash

set -u

output_dir=${1}

mkdir -p ${output_dir}

counter=0;
while [[ counter -ne "3" ]]; do
  sleep 1
  echo "executing operation" >> ${output_dir}/artifact
  ((counter++))
done