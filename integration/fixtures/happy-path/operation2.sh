#!/usr/bin/env bash

counter=0;
while [[ counter -ne "3" ]]; do
  sleep 1
  echo "executing operation 2"
  ((counter++))
done

echo "operation 2 stderr message" 1>&2