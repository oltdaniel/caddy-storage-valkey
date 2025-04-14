#!/bin/bash

# Source: https://github.com/valkey-io/valkey/blob/unstable/utils/gen-test-certs.sh

# Generate some test certificates which are used by the regression test suite:
#
#   tests/ca.{crt,key}          Self signed CA certificate.
#   tests/valkey.{crt,key}       A certificate with no key usage/policy restrictions.
#   tests/client.{crt,key}      A certificate restricted for SSL client usage.
#   tests/server.{crt,key}      A certificate restricted for SSL server usage.
#   tests/valkey.dh              DH Params file.

generate_cert() {
    local name=$1
    local cn="$2"
    local opts="$3"

    local keyfile=tests/${name}.key
    local certfile=tests/${name}.crt

    [ -f $keyfile ] || openssl genrsa -out $keyfile 2048
    openssl req \
        -new -sha256 \
        -subj "/O=Valkey Test/CN=$cn" \
        -key $keyfile | \
        openssl x509 \
            -req -sha256 \
            -CA tests/ca.crt \
            -CAkey tests/ca.key \
            -CAserial tests/ca.txt \
            -CAcreateserial \
            -days 365 \
            $opts \
            -out $certfile
}

mkdir -p tests
[ -f tests/ca.key ] || openssl genrsa -out tests/ca.key 4096
openssl req \
    -x509 -new -nodes -sha256 \
    -key tests/ca.key \
    -days 3650 \
    -subj '/O=Valkey Test/CN=Certificate Authority' \
    -out tests/ca.crt

cat > tests/openssl.cnf <<_END_
[ server_cert ]
keyUsage = digitalSignature, keyEncipherment
nsCertType = server
subjectAltName = DNS:localhost,IP:127.0.0.1

[ client_cert ]
keyUsage = digitalSignature, keyEncipherment
nsCertType = client
_END_

generate_cert server "Server-only" "-extfile tests/openssl.cnf -extensions server_cert"
generate_cert client "Client-only" "-extfile tests/openssl.cnf -extensions client_cert"
#generate_cert valkey "Generic-cert"

[ -f tests/valkey.dh ] || openssl dhparam -out tests/valkey.dh 2048