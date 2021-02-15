package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)


func main() {
	usage := `** Haproxy consul backend slot plugin **
This plugin outputs the length of the servers in each backend * 2
usage call:  ./haproxy-slot-calculator <servicename>
`

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("%s", usage))
		os.Exit(1)
	}

	env_var := os.Getenv("CONSUL_TEMPLATE_OPTS")
	consul_host := strings.Split(env_var, "=")[1]
	clustername := os.Getenv("ECS_CLUSTER")

	var defaultResponse string = fmt.Sprintf("server-template %s 5 _%s._tcp.service.consul resolvers consul check inter 5s", os.Args[1], os.Args[1])
	serviceResponse, err := request(consul_host, os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf(defaultResponse))
		os.Exit(0)
	}
	if !isRenderable(serviceResponse, clustername) {
		fmt.Fprintln(os.Stdout, fmt.Sprintf(""))
		os.Exit(0)
	}

	finalVal := PersistedCounter(os.Args[1], len(serviceResponse))
	fmt.Fprintln(os.Stdout, fmt.Sprintf(makeResponse(os.Args[1], finalVal)))
	os.Exit(0)
}

func PersistedCounter(servicename string, len int) string {
	var strlen string
	if len < 5 {
		strlen = strconv.Itoa(5)
	} else {
		strlen = strconv.Itoa(20)
	}
	return strlen
}

func request(host string, service string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s/v1/catalog/service/%s", host, service)
	spaceClient := http.Client{
		Timeout: time.Second * 5, // 5 secs
	}
	// Just for the sake of returning an empty datatype
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := spaceClient.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var lAnything []map[string]interface{}
	err = json.Unmarshal(body, &lAnything)
	return lAnything, nil
}

func isRenderable(servicedef []map[string]interface{}, clustername string) bool {
	tagmap := servicedef[0]["ServiceTags"]
	tagstring := fmt.Sprintf("%v", tagmap)
	tagtrim := strings.Trim(strings.Trim(tagstring, "["), "]")
	tagslice := strings.Split(tagtrim, " ")
	for _, v := range tagslice {
		if v == clustername {
			return true
		}
	}
	return false
}

func makeResponse(servicename string, slot string) string {
	return fmt.Sprintf("server-template %s %s _%s._tcp.service.consul resolvers consul check inter 5s", servicename, slot, servicename)
}