#!/usr/bin/env bash
set -euxo pipefail

# Generate CA root key & certificate

openssl genrsa -out ca.key 2048

openssl req -x509 -new -nodes \
        -key ca.key -subj "/CN=ExampleCA/C=US/L=NY" \
        -days 1825 -out ca.crt

## Generate server TLS key & certificate

openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr -config server-csr.conf

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
        -CAcreateserial -out server.crt -days 1825 \
        -extensions v3_ext -extfile server-csr.conf

# Generate Alice's TLS key & certificate

openssl genrsa -out alice.key 2048

openssl req -new -key alice.key -out alice.csr -config alice-csr.conf

openssl x509 -req -in alice.csr -CA ca.crt -CAkey ca.key \
        -CAcreateserial -out alice.crt -days 1825 \
        -extensions v3_ext -extfile alice-csr.conf

# Generate Bob's TLS key & certificate

openssl genrsa -out bob.key 2048

openssl req -new -key bob.key -out bob.csr -config bob-csr.conf

openssl x509 -req -in bob.csr -CA ca.crt -CAkey ca.key \
        -CAcreateserial -out bob.crt -days 1825 \
        -extensions v3_ext -extfile bob-csr.conf

# Cleanup

rm *.csr
