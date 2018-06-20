#!/usr/bin/env bash

counter=0;
while [[ counter -ne "3" ]]; do
  sleep 1
  echo "executing operation 1"
  ((counter++))
done

echo "operation 1 stderr message" 1>&2