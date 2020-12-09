# Steps to run this script:

1. Connect to RedHat VPN.
2. Authenticate yourself to RedHat customer portal - https://access.redhat.com/terms-based-registry/
3. Run `./oc-get-index-redhat-operators.sh`- index databases would download locally.
4. Run `./run-report.sh` - A file with name `report.txt` should have been generated.
# indexdump

This command line program dumps out the 'certified' annotation for each operator it finds in an index.db.

The program runs the following SQL:
SELECT name, csv FROM operatorbundle where csv is not null order by name

It then unmarshals the CSV into a struct where we can pick and choose what we
want to print out, in this case we print out the following:

operatorname, certified=true/false

You run the program as follows passing in the sqlite index.db file:
go run indexdump.go ../tiger/stagetools/index.db.4.5.stage

Column headings are:
name, csvStruct.Spec.Version, certified, createdAt, companyName, sourceDescription, repo, ocpVersion, sdkVersion, operatorType

Output looks like:

Operator [serverless-operator.v1.10.0] [certified=false]
Operator [serverless-operator.v1.7.2] [certified=false]
Operator [service-binding-operator.v0.3.0] [certified=false]

## OperatorHub Indexes

Here is a link to the various marketplace operator indexes from where you can pull
data:
https://github.com/operator-framework/operator-marketplace/tree/master/defaults

To download indexes from that location:

docker login https://registry.redhat.io

Then run the  oc-get-index-redhat-operators.sh script to download the index
files.

also...
docker login registry.connect.redhat.com
Username: jemccorm@redhat.com


https://github.com/operator-framework/operator-lifecycle-manager/blob/master/deploy/upstream/manifests/0.16.1/0000_50_olm_17-upstream-operators.catalogsource.yaml#L10

docker login quay.io
quay.io/operatorhubio/catalog:latest

NOTE:  that image is built as follows...

```
We use a feature in quay to build it

We configured that repository to have a job that runs whenever a change is committed to the upstream-community-operators folder in the community-operators repo

It points to this dockerfile https://github.com/operator-framework/community-operators/blob/master/upstream.Dockerfile

Which basically just uses an image from the operator-registry to build the catalog
```
