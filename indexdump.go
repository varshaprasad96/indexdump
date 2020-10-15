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

	pathToIndexFile := args[0]
	db, err := sql.Open("sqlite3", pathToIndexFile)
	if err != nil {
		panic(err)
	}

	dump(db)
}

func dump(db *sql.DB) {
	row, err := db.Query("SELECT name, csv FROM operatorbundle where csv is not null order by name")
	if err != nil {
		panic(err)
	}
	var csvStruct v1alpha1.ClusterServiceVersion

	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var name string
		var csv string
		row.Scan(&name, &csv)
		//fmt.Println("Operator: ", name, " csv ", csv)
		err := json.Unmarshal([]byte(csv), &csvStruct)
		if err != nil {
			fmt.Printf("error unmarshalling csv %s\n", err.Error())
		}
		//fmt.Printf("Operator: %s\n", name)
		/**
		for k, v := range csvStruct.ObjectMeta.Labels {
			fmt.Printf("Label [%s] [%s]\n", k, v)
		}
		*/
		for k, v := range csvStruct.ObjectMeta.Annotations {
			if k == "certified" {
				fmt.Printf("Operator [%s] [%s=%s]\n", name, k, v)
			}
		}
	}
}
