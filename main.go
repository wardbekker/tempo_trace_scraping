package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Trace struct {
	TraceId           string `json:"traceId"`
	RootServiceName   string `json:"rootServiceName"`
	RootTraceName     string `json:"rootTraceName"`
	StartTimeUnixNano string `json:"startTimeUnixNano"`
	DurationMs        int    `json:"durationMs"`
}

type TraceMetrics struct {
	InspectedTraces int    `json:"inspectedTraces"`
	InspectedBytes  string `json:"inspectedBytes"`
	InspectedBlocks int    `json:"inspectedBlocks"`
}

type TracesResponse struct {
	Traces       []Trace      `json:"traces"`
	TraceMetrics TraceMetrics `json:"metrics"`
}

func main() {

	// get cli args
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./main username password [endpoint]")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	endpoint := "https://tempo-us-central1.grafana.net"

	if len(os.Args) == 4 {
		endpoint = os.Args[3]
		fmt.Println("Using endpoint: " + endpoint)
	}

	// https://grafana.com/docs/tempo/latest/api_docs/
	// do a http get request to the grafana api

	req, _ := http.NewRequest("GET", endpoint+"/tempo/api/search?limit=1000", nil)
	req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
	cli := &http.Client{}

	resp, _ := cli.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)

	// deserialize the response as json
	var tr TracesResponse

	err := json.Unmarshal(body, &tr)

	if err != nil {
		fmt.Println(err)
	}

	println("retrieved " + fmt.Sprintf("%d", len(tr.Traces)) + " traces")

	// loop over the traces
	for _, trace := range tr.Traces {
		// do a get request to get the trace Data
		req, _ := http.NewRequest("GET", endpoint+"/tempo/api/traces/"+trace.TraceId, nil)
		req.Header.Add("Authorization", "Basic "+basicAuth(username, password))
		cli = &http.Client{}
		resp, _ = cli.Do(req)

		body, _ = ioutil.ReadAll(resp.Body)

		// write body to a new file
		ioutil.WriteFile("trace_"+trace.TraceId+".json", body, 0644)

		// sleep for a few ms to not overload the server
		time.Sleep(time.Millisecond * 50)
		println("writing trace: " + trace.TraceId)
	}
	println("done")
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
