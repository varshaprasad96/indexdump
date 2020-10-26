package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {
	//URL := "https://raw.githubusercontent.com/jmccormick2001/operator-sdk/master/go.mod"
	//URL := "https://github.com/jmccormick2001/operator-sdk/blob/master/go.mod"

	// Open the file
	file, err := os.Open("report.txt.sorted")
	if err != nil {
		fmt.Printf("Couldn't open the csv file", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "|")
		repoURL := cols[6]
		//fmt.Println(repoURL)
		sdkVersion, found, operatorType := getSDKVersion(repoURL)
		if !found {
			sdkVersion, found, operatorType = getAnsibleHelmVersion(repoURL)
		}
		fmt.Printf("[%s] [%s] [%s]\n", cols[0], sdkVersion, operatorType)

	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

//		URL := repoURL + "/blob/master/go.mod"
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
