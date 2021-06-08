#!/usr/bin/env bash

export BASE64ENCODED_METAL_PROVIDER_CREDS=$(base64 credentials.txt | tr -d "\n")
sed "s/BASE64ENCODED_METAL_PROVIDER_CREDS/$BASE64ENCODED_METAL_PROVIDER_CREDS/g" provider.yaml | kubectl create -f -