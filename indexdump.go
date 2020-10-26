package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
		sdkVersion, found, operatorType := getSDKVersion(repo)
		if !found {
			sdkVersion, found, operatorType = getAnsibleHelmVersion(repo)
		}
		fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n", name, csvStruct.Spec.Version, certified, createdAt, companyName, sourceDescription, repo, ocpVersion, sdkVersion, operatorType)
	}
}

func getSDKVersion(inURL string) (sdkVersion string, found bool, operatorType string) {
	//replace github.com with raw.githubusercontent.com
	URL := strings.Replace(inURL, "github.com", "raw.githubusercontent.com", 1)
	URL = URL + "/master/go.mod"
	//URL := "https://raw.githubusercontent.com/3scale/3scale-operator/master/go.mod"
	//	fmt.Printf("trying URL %s\n", URL)
	response, err := http.Get(URL) //use package "net/http"

	if err != nil {
		//fmt.Println("go.mod not found " + err.Error())
		return "", false, ""
	}

	defer response.Body.Close()

	// Copy data from the response to standard output
	body, err1 := ioutil.ReadAll(response.Body) //use package "io" and "os"
	if err != nil {
		fmt.Println(err1)
		return "", false, ""
	}

	//	fmt.Println("Number of bytes copied to STDOUT:", n)
	temp := strings.Split(string(body), "\n")
	for i := 0; i < len(temp); i++ {
		if strings.Contains(temp[i], "operator-sdk") &&
			!strings.Contains(temp[i], "=>") &&
			!strings.Contains(temp[i], "replace") {
			//fmt.Printf("%s\n", temp[i])
			sdkVersion := strings.Split(strings.TrimSpace(temp[i]), " ")
			if len(sdkVersion) > 1 {
				//fmt.Printf("version [%s]\n", sdkVersion[1])
				return sdkVersion[1], true, "golang"
			}
		}
	}
	return "", false, ""

}

//		URL := repoURL + "/blob/master/build/Dockerfile"
func getAnsibleHelmVersion(inURL string) (sdkVersion string, found bool, operatorType string) {
	//replace github.com with raw.githubusercontent.com
	URL := strings.Replace(inURL, "github.com", "raw.githubusercontent.com", 1)
	URL = URL + "/master/build/Dockerfile"
	//URL := "https://raw.githubusercontent.com/3scale/3scale-operator/master/go.mod"
	//	fmt.Printf("trying URL %s\n", URL)
	response, err := http.Get(URL)

	if err != nil {
		//fmt.Println("build/Dockerfile not found " + err.Error())
		return "", false, ""
	}

	defer response.Body.Close()

	// Copy data from the response to standard output
	body, err1 := ioutil.ReadAll(response.Body) //use package "io" and "os"
	if err != nil {
		fmt.Println(err1)
		return "", false, ""
	}

	//	fmt.Println("Number of bytes copied to STDOUT:", n)
	temp := strings.Split(string(body), "\n")
	for i := 0; i < len(temp); i++ {
		if strings.Contains(temp[i], "ansible-operator") &&
			strings.Contains(temp[i], "operator-framework") {
			//fmt.Printf("%s\n", temp[i])
			sdkVersion := strings.Split(strings.TrimSpace(temp[i]), " ")
			if len(sdkVersion) > 1 {
				//fmt.Printf("version [%s]\n", sdkVersion[1])
				return getSDKVersionFromImage(sdkVersion[1]), true, "ansible"
			}
		} else if strings.Contains(temp[i], "helm-operator") &&
			strings.Contains(temp[i], "operator-framework") {
			sdkVersion := strings.Split(strings.TrimSpace(temp[i]), " ")
			if len(sdkVersion) > 1 {
				//fmt.Printf("version [%s]\n", sdkVersion[1])
				return getSDKVersionFromImage(sdkVersion[1]), true, "helm"
			}
		}
	}
	return "", false, ""

}

func getSDKVersionFromImage(input string) (output string) {
	result := strings.Split(input, ":")
	l := len(result)
	if l > 0 {
		return result[l-1]
	}
	return ""
}
