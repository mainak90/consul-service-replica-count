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
	var check bool
	_ ,ok := os.LookupEnv("TIMER")
	if !ok {
		check = false
	} else {
		check = true
	}
	consul_host := strings.Split(os.Getenv("CONSUL_TEMPLATE_OPTS"), "=")[1]
	finalVal := PersistedCounter(consul_host, os.Args[1], check)
	iszero := IsZeroOfUnderlyingType(finalVal)
	if (iszero) {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", finalVal))
		os.Exit(0)
	}
}

func PersistedCounter(host string, servicename string, check bool) string {
	result := make(map[string]string)
	var count int
	servicesPath,ok := os.LookupEnv("JSON_PATH")
	if !ok {
		servicesPath = "/etc/consul-template/services.json"
	}
	if _, err := os.Stat(servicesPath); os.IsNotExist(err) {
		os.Create(servicesPath)
	}
	var strlen string
	jsonFile, err := os.Open(servicesPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "10"
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &result)
	if _, ok := result[servicename]; ok {
		if (result[servicename] == "20") {
			return result[servicename]
		} else {
			if (check) {
				//fmt.Println("Checktime is enabled and service exists on services json")
				willcheck := CheckTimed(host, servicename)
				if (willcheck) {
					//fmt.Println("Getting service count via request but service was pre-existing in service.json")
					count = request(host, servicename)
					StoreExactCount(servicename, count)
				} else {
					//fmt.Println("Getting service count via services.json as check interval is not passed")
					count, _ = strconv.Atoi(result[servicename])
				}
			} else {
				//fmt.Println("Checktime is disabled so directly querying the service but service exists in service.json")
				count = request(host, servicename)
				StoreExactCount(servicename, count)
			}
			if count <= 10 {
				strlen = "10"
			} else {
				strlen = "20"
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
	}
	if (check) {
		//fmt.Println("Seems like a new service and checktime is enabled")
		_ = CheckTimed(host, servicename)
		count = request(host, servicename)
		StoreExactCount(servicename, count)
	} else {
		//fmt.Println("Seems like a new service and checktime is disabled")
		count = request(host, servicename)
		StoreExactCount(servicename, count)
	}
	if count < 10 {
		strlen = "10"
	} else {
		strlen = "20"
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

func request(host string, service string) int {
	url := fmt.Sprintf("http://%s/v1/catalog/service/%s", host, service)
	defaultcount := 10
	spaceClient := http.Client{
		Timeout: time.Second * 5, // 5 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return defaultcount
	}
	res, err := spaceClient.Do(req)
	if err != nil {
		return defaultcount
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return defaultcount
	}
	var lAnything []map[string]interface{}
	err = json.Unmarshal(body, &lAnything)
	return len(lAnything)
}

func CheckTimed(host string, servicename string) bool {
	if _, err := os.Stat("/etc/consul-template/timestamps.json"); os.IsNotExist(err) {
		os.Create("/etc/consul-template/timestamps.json")
	}
	timestamp := make(map[string]int64)
	jsonFile, err := os.Open("/etc/consul-template/timestamps.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return true
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &timestamp)
	currcount := GetExactCount(host, servicename)
	//fmt.Println("Current count is : ", strconv.Itoa(currcount))
	if _, ok := timestamp[servicename]; ok {
		if currcount < 5 {
			//fmt.Println("Looks like the application has 4 or less replicas")
			if time.Now().Add(60 * -time.Second).Unix() > timestamp[servicename] {
				timestamp[servicename] = time.Now().Unix()
				dataBytes, _ := json.Marshal(timestamp)
				_ = ioutil.WriteFile("/etc/consul-template/timestamps.json", dataBytes, 0777)
				return true
			} else {
				return false
			}
		} else {
			//fmt.Println("Application has more than 4 replicas, so it will always query from http request")
			return true
		}
	}
	timestamp[servicename] = time.Now().Unix()
	dataBytes, _ := json.Marshal(timestamp)
	_ = ioutil.WriteFile("/etc/consul-template/timestamps.json", dataBytes, 0777)
	return true
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	if reflect.DeepEqual("0", x) {
		return true
	} else {
		return false
	}
}

func StoreExactCount(servicename string, count int) {
	if _, err := os.Stat("/etc/consul-template/servicecounters.json"); os.IsNotExist(err) {
		os.Create("/etc/consul-template/servicecounters.json")
	}
	counter := make(map[string]int)
	jsonFile, _ := os.Open("/etc/consul-template/servicecounters.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &counter)
	counter[servicename] = count
	dataBytes, _ := json.Marshal(counter)
	_ = ioutil.WriteFile("/etc/consul-template/servicecounters.json", dataBytes, 0777)
}

func GetExactCount(host string, servicename string) int {
	if _, err := os.Stat("/etc/consul-template/servicecounters.json"); os.IsNotExist(err) {
		count := request(host, servicename)
		StoreExactCount(servicename, count)
		return count
	}
	countered := make(map[string]int)
	jsonFile, _ := os.Open("/etc/consul-template/servicecounters.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal([]byte(byteValue), &countered)
	if _, ok := countered[servicename]; ok {
		return countered[servicename]
	}
	count := request(host, servicename)
	StoreExactCount(servicename, count)
	return count
}
