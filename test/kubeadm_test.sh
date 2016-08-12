#!/bin/sh

KUBEADM="${KUBEADM:-kubeadm}"

# XXX turn this into golang tests that doesn't use GLOBAL MUTABLE STATE ON THE
# HOST :( :( :(

# cleanup
rm -f /etc/kubernetes/kubelet-bootstrap.json

$KUBEADM manual bootstrap master \
    --api-dns-name="mycoolkubernetescluster.io"

# TODO test that:
# 1. an /etc/kubernetes/kubelet-bootstrap.json file has been written
# 2. some static pods have been written
# 3. the initial ca-cert file is generated (unless the user provides ca-cert
#    file and key)

rm -f /etc/kubernetes/kubelet-bootstrap.json

$KUBEADM manual bootstrap node \
    --api-servers="http://127.0.0.1:8080" \
    --ca-cert-file="/etc/kubernetes/ca.pem"


