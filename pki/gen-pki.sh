#!/usr/bin/env bash
set -euxo pipefail

# Generate CA root key & certificate

openssl genrsa -out ca.key 2048

openssl req -x509 -new -nodes -sha256 \
        -addext basicConstraints=critical,CA:TRUE,pathlen:0 \
        -addext keyUsage=cRLSign,keyCertSign \
        -subj "/CN=ExampleCA/C=AU/ST=NSW/L=Sydney" \
        -days 1825 \
        -key ca.key \
        -out ca.crt

## Generate server TLS key & certificate

openssl genrsa -out server.key 2048

openssl req -new -key server.key -out server.csr -config server-csr.conf

openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -days 1825 \
        -extensions v3_ext -extfile server-csr.conf \
        -in server.csr -out server.crt

# Generate Alice's TLS key & certificate

openssl genrsa -out alice.key 2048

openssl req -new -key alice.key -out alice.csr -config alice-csr.conf

openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -days 1825 \
        -extensions v3_ext -extfile alice-csr.conf \
        -in alice.csr -out alice.crt

# Generate Bob's TLS key & certificate

openssl genrsa -out bob.key 2048

openssl req -new -key bob.key -out bob.csr -config bob-csr.conf

openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -days 1825 \
        -extensions v3_ext -extfile bob-csr.conf \
        -in bob.csr -out bob.crt

# Cleanup

rm ./*.csr
