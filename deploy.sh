#!/bin/bash

docker build -t tvilarim/eng-atestados-go . &&

docker push tvilarim/eng-atestados-go:latest &&

kubectl apply -f k8s/deployment.yaml