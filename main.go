package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)


func main() {
	version := "1.0.0"
	usage := `** Haproxy consul backend slot plugin **
This plugin outputs the length of the servers in each backend * 2
usage call:  ./haproxy-slot-calculator <servicename>
`

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", usage))
		os.Exit(1)
	}

	if os.Args[1] == "-version" {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", version))
		os.Exit(1)
	}

	env_var := os.Getenv("CONSUL_TEMPLATE_OPTS")
	consul_host := strings.Split(env_var, "=")[1]
	url := fmt.Sprintf("http://%s/v1/catalog/service/%s", consul_host, os.Args[1])

	spaceClient := http.Client{
		Timeout: time.Second * 5, // 5 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", err))
		os.Exit(1)
	}
	res, err := spaceClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", err))
		os.Exit(1)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", err))
		os.Exit(1)
	}
	var lAnything []map[string]interface{}
	err = json.Unmarshal(body, &lAnything)
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", err))
		os.Exit(1)
	}
	finalVal := PersistedCounter(os.Args[1], len(lAnything))
	iszero := IsZeroOfUnderlyingType(finalVal)
	if (iszero) {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", finalVal))
		os.Exit(0)
	}
}

func PersistedCounter(servicename string, len int) string {
	result := make(map[string]string)
	servicesPath,ok := os.LookupEnv("JSON_PATH")
	if !ok {
		servicesPath = "/etc/consul-template/services.json"
	}
	if _, err := os.Stat(servicesPath); os.IsNotExist(err) {
		os.Create(servicesPath)
	}
	var strlen string
	if len < 10 {
		strlen = strconv.Itoa(10)
	} else {
		strlen = strconv.Itoa(20)
	}
	jsonFile, err := os.Open(servicesPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "10"
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &result)
	if _, ok := result[servicename]; ok {
		if (result[servicename] == strlen || result[servicename] == "20") {
			return result[servicename]
		} else {
			result[servicename] = strlen
			dataBytes, err := json.Marshal(result)
			if err != nil {
				return "10"
			}
			err = ioutil.WriteFile(servicesPath, dataBytes, 0777)
			if err != nil {
				return "10"
			}
			return strlen
		}
	}
	result[servicename] = strlen
	dataBytes, err := json.Marshal(result)
	if err != nil {
		return "10"
	}
	err = ioutil.WriteFile(servicesPath, dataBytes, 0777)
	if err != nil {
		return "10"
	}
	return strlen
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	if reflect.DeepEqual("0", x) {
		return true
	} else {
		return false
	}
}
