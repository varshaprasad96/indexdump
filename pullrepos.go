package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"os"
	"sort"
	"strings"
)

type ReportColumns struct {
	Operator          string
	Version           string
	Certified         string
	CreatedAt         string
	Company           string
	Repo              string
	OCPVersion        string
	SDKVersion        string
	OperatorType      string
	SourceRedhat      string
	SourceCommunity   string
	SourceMarketplace string
	SourceCertified   string
	SourceOperatorHub string
}

var ReportMap map[string]ReportColumns

type Inputs struct {
	Path    string
	Source  string
	Version string
}

var InputsList []Inputs

func main() {
	ReportMap = make(map[string]ReportColumns)
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("path is a required argument\n")
		os.Exit(1)
	}
	InputsList = make([]Inputs, 0)
	for i := 0; i < len(args); i++ {
		//	fmt.Printf("arg %s\n", args[i])
		v := strings.Split(args[i], ":")
		input := Inputs{
			Path:    v[0],
			Source:  v[1],
			Version: v[2],
		}
		InputsList = append(InputsList, input)
	}

	for i := 0; i < len(InputsList); i++ {
		db, err := sql.Open("sqlite3", InputsList[i].Path)
		if err != nil {
			panic(err)
		}

		dump(db, InputsList[i].Source, InputsList[i].Version)
	}

	printReport()

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

		repo := csvStruct.ObjectMeta.Annotations["repository"]

		if repo != "" {
			f, ok := ReportMap[name]
			if ok {
				//update the entry's source columns
				//fmt.Printf("Jeff - update an entry %s\n", name)
			} else {
				ReportMap[name] = ReportColumns{
					Operator: name,
					Repo:     repo,
				}
				f = ReportMap[name]
			}
			ReportMap[name] = f
		}

	}
}

func printReport() {
	keys := make([]string, 0, len(ReportMap))
	for k := range ReportMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// print the 1st row which acts as the spreadsheet header
	for _, k := range keys {
		f := ReportMap[k]
		fmt.Printf("git clone %s\n", f.Repo)
	}
}
