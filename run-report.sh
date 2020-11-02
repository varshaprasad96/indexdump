#!/bin/bash
FILE=report.txt
SORTEDFILE=report.sorted.txt
if test -f "$SORTEDFILE"; then
    echo "$SORTEDFILE exists."
    mv $SORTEDFILE $SORTEDFILE.previous
fi
echo Column headers are as follows:
echo operator,version,certified,created,company,repos,ocpversion,sdkversion,operatortype,source-redhat,source-community,source-marketplace,source-certified,source-operatorhub,source-prod,channel

go run indexdump.go \
"index.db.4.6.prod:prod:4.6" \
"index.db.4.6.redhat-operators:redhat:4.6" \
"index.db.4.6.community-operators:community:4.6" \
"index.db.4.6.redhat-marketplace-operators:marketplace:4.6" \
"index.db.4.6.certified-operators:certified:4.6" \
"index.db.operatorhub.io:operatorhub:4.6" > $FILE

exit
sort $FILE > $FILE.sorted

echo $FILE.sorted file was created
rm $FILE
