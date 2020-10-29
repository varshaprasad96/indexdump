#!/bin/bash
FILE=repos.txt

go run pullrepos.go \
"index.db.4.6.redhat-operators:redhat:4.6" \
"index.db.4.6.community-operators:community:4.6" \
"index.db.4.6.redhat-marketplace-operators:marketplace:4.6" \
"index.db.4.6.certified-operators:certified:4.6" \
"index.db.operatorhub.io:operatorhub:4.6" > $FILE

mkdir repos
cd repos
source ../$FILE
