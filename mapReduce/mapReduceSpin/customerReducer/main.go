package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type REQ struct {
	BucketName   string `json:"bucketName"`
	Key          string `json:"key"`
	OutputBucket string `json:"outputBucket"`
}

type ValueData struct {
	Value int `json:"value"`
}

type apiResp struct {
	Data []ValueData `json:"data"`
}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		reduce(w, r)
	})
}

func reduce(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	var cc REQ
	var err error
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(body, &cc); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	baseURL := "http://localhost:8085/getFilesFromBucketWitPrefix"

	params := url.Values{}
	params.Add("bucketName", cc.BucketName)
	params.Add("key", "key/"+cc.Key)
	params.Add("formatJson", "true")

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		http.Error(w, "Error assembling QueryUrl", http.StatusBadRequest)
		return
	}

	// Attach the encoded query parameters to the URL
	parsedURL.RawQuery = params.Encode()

	response, err := spinhttp.Get(parsedURL.String())
	if err != nil {
		http.Error(w, "Error cant get data to reduce", http.StatusInternalServerError)
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var lines apiResp
	if err := json.Unmarshal(resp, &lines); err != nil {
		http.Error(w, "Cant Unmarshal resp", http.StatusInternalServerError)
	}

	var amount int

	for _, line := range lines.Data {
		amount += line.Value
	}

	baseURL = "http://localhost:8085/putObject"

	params = url.Values{}
	params.Add("bucketName", cc.OutputBucket)
	params.Add("key", cc.Key)

	parsedURL, err = url.Parse(baseURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}

	// Attach the encoded query parameters to the URL
	parsedURL.RawQuery = params.Encode()

	// The fully constructed URL
	fullURL := parsedURL.String()

	outputData, err := json.Marshal(map[string]int{cc.Key: amount})
	if err != nil {
		log.Fatalln(err)
	}

	_, err = spinhttp.Post(fullURL, "application/json", bytes.NewReader(outputData))
	if err != nil {
		http.Error(w, "Error reading putting Object from input with the key : "+cc.Key, http.StatusInternalServerError)
		return
	}

	endTime := time.Now()
	result := map[string]interface{}{
		"key":  cc.Key,
		"time": endTime.Sub(startTime).Seconds(),
	}

	responseBytes, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

func main() {}
