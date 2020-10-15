# indexdump

This command line program dumps out the 'certified' annotation for each operator it finds in an index.db.

The program runs the following SQL:
SELECT name, csv FROM operatorbundle where csv is not null order by name

It then unmarshals the CSV into a struct where we can pick and choose what we
want to print out, in this case we print out the following:

operatorname, certified=true/false

You run the program as follows passing in the sqlite index.db file:
go run indexdump.go ../tiger/stagetools/index.db.4.5.stage

Output looks like:

Operator [serverless-operator.v1.10.0] [certified=false]
Operator [serverless-operator.v1.7.2] [certified=false]
Operator [service-binding-operator.v0.3.0] [certified=false]
Operator [service-registry-operator.v1.0.2] [certified=false]
Operator [servicemeshoperator.v1.1.10] [certified=false]
Operator [web-terminal.v1.0.1] [certified=false]
