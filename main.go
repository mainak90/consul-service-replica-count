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
	finalVal := strconv.Itoa(len(lAnything) * 2)
	iszero := IsZeroOfUnderlyingType(finalVal)
	if (iszero) {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("10"))
		os.Exit(0)
	} else {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s", finalVal))
		os.Exit(0)
	}
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	if reflect.DeepEqual("0", x) {
		return true
	} else {
		return false
	}
}
