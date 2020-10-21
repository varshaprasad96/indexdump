#!/bin/bash
FILE=report.txt
SORTEDFILE=report.sorted.txt
if test -f "$SORTEDFILE"; then
    echo "$SORTEDFILE exists."
    mv $SORTEDFILE $SORTEDFILE.previous
fi
echo "looking at redhat-operators..."
go run indexdump.go index.db.4.6.redhat-operators redhat >> $FILE
echo "looking at community-operators..."
go run indexdump.go index.db.4.6.community-operators community >> $FILE
echo "looking at redhat-marketplace-operators..."
go run indexdump.go index.db.4.6.redhat-marketplace-operators marketplace >> $FILE
echo "looking at certified-operators..."
go run indexdump.go index.db.4.6.certified-operators certified >> $FILE
echo "looking at operatorhub.io operators..."
go run indexdump.go index.db.operatorhub.io operatorhub >> $FILE

sort $FILE > $FILE.sorted

echo $FILE.sorted file was created
rm $FILE
