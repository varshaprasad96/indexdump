package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("path is a required argument\n")
		os.Exit(1)
	}
	if len(args) != 3 {
		fmt.Printf("path, sourceDescription, ocpversion are required parameters\n")
		os.Exit(1)
	}

	pathToIndexFile := args[0]
	db, err := sql.Open("sqlite3", pathToIndexFile)
	if err != nil {
		panic(err)
	}

	sourceDescription := args[1]
	ocpVersion := args[2]

	dump(db, sourceDescription, ocpVersion)
}

func dump(db *sql.DB, sourceDescription, ocpVersion string) {
	row, err := db.Query("SELECT name, csv FROM operatorbundle where csv is not null order by name")
	if err != nil {
		panic(err)
	}
	var csvStruct v1alpha1.ClusterServiceVersion

	//fmt.Println("operator, version, certified, createdAt, company, source, repo, ocpversion")
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var csv string
		row.Scan(&name, &csv)
		err := json.Unmarshal([]byte(csv), &csvStruct)
		if err != nil {
			fmt.Printf("error unmarshalling csv %s\n", err.Error())
		}
		//fmt.Printf("%s\n", csv)
		//os.Exit(1)
		/**
		for k, v := range csvStruct.ObjectMeta.Labels {
			fmt.Printf("Label [%s] [%s]\n", k, v)
		}
		*/
		/**
		for k, v := range csvStruct.ObjectMeta.Annotations {
			if k == "certified" {
				//fmt.Printf("Operator [%s] [%s=%s]\n", name, k, v)
				certified = v
				break
			}
		}
		*/
		certified := csvStruct.ObjectMeta.Annotations["certified"]
		repo := csvStruct.ObjectMeta.Annotations["repository"]
		createdAt := csvStruct.ObjectMeta.Annotations["createdAt"]
		companyName := csvStruct.Spec.Provider.Name
		fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s\n", name, csvStruct.Spec.Version, certified, createdAt, companyName, sourceDescription, repo, ocpVersion)
	}
}
