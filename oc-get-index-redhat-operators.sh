#!/bin/bash
set -x
oc image extract registry.redhat.io/redhat/redhat-operator-index:v4.6 --file=/database/index.db --filter-by-os='linux/amd64'
mv index.db index.db.4.6.redhat-operators
oc image extract registry.redhat.io/redhat/certified-operator-index:v4.6 --file=/database/index.db --filter-by-os='linux/amd64'
mv index.db index.db.4.6.certified-operators
oc image extract registry.redhat.io/redhat/community-operator-index:v4.6 --file=/database/index.db --filter-by-os='linux/amd64'
mv index.db index.db.4.6.community-operators
oc image extract registry.redhat.io/redhat/redhat-marketplace-index:v4.6 --file=/database/index.db --filter-by-os='linux/amd64'
mv index.db index.db.4.6.redhat-marketplace-operators
oc image extract quay.io/operatorhubio/catalog:latest --file=/database/index.db
mv index.db index.db.operatorhub.io
oc image extract registry-proxy.engineering.redhat.com/rh-osbs/iib-pub:v4.6 --file=/database/index.db --filter-by-os='linux/amd64'
mv index.db index.db.4.6.prod