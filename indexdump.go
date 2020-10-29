package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/operator-framework/api/pkg/manifests"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"sort"
	"strings"
)

const source_redhat = "redhat"
const source_community = "community"
const source_marketplace = "marketplace"
const source_certified = "certified"
const source_operatorhub = "operatorhub"

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
	Channel           string
	DefaultChannel    string
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
	//fmt.Printf("inputsList %+v\n", InputsList)

	for i := 0; i < len(InputsList); i++ {
		//fmt.Printf("opening %s\n", InputsList[i].Path)
		db, err := sql.Open("sqlite3", InputsList[i].Path)
		if err != nil {
			panic(err)
		}

		dump(db, InputsList[i].Source, InputsList[i].Version)

		//TODO REMOVE THIS if stmt
		/**
		if i == 0 {
			fmt.Printf("jeff breaking out after 1\n")
			break
		}
		*/
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

		certified := csvStruct.ObjectMeta.Annotations["certified"]

		repo := csvStruct.ObjectMeta.Annotations["repository"]
		exists, repoPath := repoExists(repo)
		channel := "unknown"
		defaultChannel := "unknown"
		if exists {
			channel, defaultChannel, err = getChannel(repoPath)
		}

		createdAt := csvStruct.ObjectMeta.Annotations["createdAt"]
		companyName := csvStruct.Spec.Provider.Name
		sdkVersion, found, operatorType := getSDKVersion(repo)
		if !found {
			sdkVersion, found, operatorType = getAnsibleHelmVersion(repo)
		}

		f, ok := ReportMap[name]
		if ok {
			//update the entry's source columns
			//fmt.Printf("Jeff - update an entry %s\n", name)
		} else {
			ReportMap[name] = ReportColumns{
				Operator:       name,
				Version:        csvStruct.Spec.Version.String(),
				Certified:      certified,
				CreatedAt:      createdAt,
				Company:        companyName,
				Repo:           repo,
				OCPVersion:     ocpVersion,
				SDKVersion:     sdkVersion,
				OperatorType:   operatorType,
				Channel:        channel,
				DefaultChannel: defaultChannel,
			}
			f = ReportMap[name]
		}
		switch sourceDescription {
		case source_redhat:
			f.SourceRedhat = "yes"
		case source_community:
			f.SourceCommunity = "yes"
		case source_marketplace:
			f.SourceMarketplace = "yes"
		case source_certified:
			f.SourceCertified = "yes"
		case source_operatorhub:
			f.SourceOperatorHub = "yes"
		}
		ReportMap[name] = f

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

func printReport() {
	keys := make([]string, 0, len(ReportMap))
	for k := range ReportMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// print the 1st row which acts as the spreadsheet header
	fmt.Println("operator|version|certified|created|company|repos|ocpversion|sdkversion|operatortype|source-redhat|source-community|source-marketplace|source-certified|source-operatorhub|channel|default-channel")
	for _, k := range keys {
		f := ReportMap[k]
		fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n",
			f.Operator,
			f.Version,
			f.Certified,
			f.CreatedAt,
			f.Company,
			f.Repo,
			f.OCPVersion,
			f.SDKVersion,
			f.OperatorType,
			f.SourceRedhat,
			f.SourceCommunity,
			f.SourceMarketplace,
			f.SourceCertified,
			f.SourceOperatorHub,
			f.Channel,
			f.DefaultChannel)
	}
}

func repoExists(repoURL string) (exists bool, path string) {

	pieces := strings.Split(repoURL, "/")
	repoName := pieces[len(pieces)-1]
	path = "repos/" + repoName
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		exists = true
	}
	//fmt.Printf("repo %s exists %t\n", repoName, exists)
	return exists, path
}

func getChannel(repoPath string) (channel, defaultChannel string, err error) {

	pattern := "annotations.yaml"
	repoPath = repoPath

	//fmt.Printf("looking for %s in %s\n", pattern, repoPath)

	//libRegEx, e := regexp.Compile("^.+\\.(LICENSE)$")
	//libRegEx, e := regexp.Compile("^.+\\.go")
	libRegEx, e := regexp.Compile(pattern)
	if e != nil {
		log.Fatal(e)
	}

	var found bool
	var pathFound string
	e = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		//		fmt.Printf("comparing to %s\n", info.Name())
		if err == nil && libRegEx.MatchString(info.Name()) {
			//println(info.Name())
			found = true
			pathFound = path
			return nil
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}
	if found {
		content, err := ioutil.ReadFile(pathFound)
		if err != nil {
			return "", "", err
		}
		lines := strings.Split(string(content), "\n")
		//operators.operatorframework.io.bundle.channel.default.v1: "alpha"
		//operators.operatorframework.io.bundle.channels.v1: "alpha"

		for i := 0; i < len(lines); i++ {
			pieces := strings.Split(lines[i], ":")
			if strings.Contains(lines[i], "bundle.channel.default") {
				defaultChannel = pieces[1]
			}
			if strings.Contains(lines[i], "bundle.channels") {
				channel = pieces[1]
			}
		}
		return channel, defaultChannel, nil
	} else {
		//check for *.package.yaml
		channel, defaultChannel = checkForPackageYaml(repoPath)
		return channel, defaultChannel, nil
	}
	return "notfound", "notfound", nil
}

func checkForPackageYaml(repoPath string) (channel string, channelDefault string) {

	//look for *.package.yaml files
	pattern := "^.+\\.package.yaml"
	//libRegEx, e := regexp.Compile("^.+\\.go")
	libRegEx, e := regexp.Compile(pattern)
	if e != nil {
		log.Fatal(e)
	}

	var found bool
	var pathFound string
	e = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && libRegEx.MatchString(info.Name()) {
			//fmt.Printf("found %s\n", path)
			//println(info.Name())
			found = true
			pathFound = path
			return nil
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}
	if found {
		content, err := ioutil.ReadFile(pathFound)
		if err != nil {
			fmt.Println(err.Error())
			return "", ""
		}
		var pm manifests.PackageManifest
		err = yaml.Unmarshal(content, &pm)
		if err != nil {
			fmt.Println(err.Error())
			return "", ""
		}

		if len(pm.Channels) == 1 {
			channel = pm.Channels[0].Name
		} else {
			for i := 0; i < len(pm.Channels); i++ {
				channel = channel + "," + pm.Channels[i].Name
			}
		}
		channelDefault = pm.DefaultChannelName
	}
	return channel, channelDefault
}
