#!/bin/bash -e

# kube-ca.sh - convenience script for generating the kubernetes
# secret for the user CA.

if [ $# -ne 3 ]
then
    echo "Missing required argument(s)"
    echo "Usage: kube-ca.sh <crtfile> <secret name> <namespace>"
    exit 1
fi

CA_CERT="$1"
SECRET_NAME="$2"
NAMESPACE="$3"

kubectl -n $NAMESPACE create secret generic $SECRET_NAME \
    --from-file=ca.crt=${CA_CERT} \
    --dry-run=client \
    -o yaml
