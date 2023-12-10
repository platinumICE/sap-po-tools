package main

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var client *http.Client

func initiateHTTPClient(options RuntimeConfiguration) {

	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 0,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   0,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
				// should not be changed
			},
		},
		Timeout: 0,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
			// this prevents Go from following redirects
		},
	}
}

func downloadGeneric(connect ConnectionOptions, request string) XIEnvelop {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/AdapterMessageMonitoring/basic?style=document", connect.Hostname), strings.NewReader(request))
	req.SetBasicAuth(connect.Username, connect.Password)
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		// noop
	case 401:
		fmt.Printf("HTTP 401: incorrect password for user %s\n", connect.Username)
		os.Exit(2)
	case 403:
		fmt.Printf("HTTP 403: incorrect password for user %s\n", connect.Username)
		os.Exit(2)
	default:
		fmt.Printf("HTTP %s: cannot read overview for host %s\n", resp.Status, connect.Hostname)
		os.Exit(3)
	}

	responseBytes, err := io.ReadAll(resp.Body)

	httpResults := new(XIEnvelop)
	err = xml.Unmarshal(responseBytes, &httpResults)

	if err != nil {
		fmt.Printf("Please verify that host [%s], username [%s] and password are correct\n", connect.Hostname, connect.Username)
		fmt.Printf("HTTP call returned: %s\n", err.Error())
		os.Exit(3)
	}

	return *httpResults
}
