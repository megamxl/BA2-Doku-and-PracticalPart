package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		getData(w, r)
	})
}

func main() {}

type REQ struct {
	BucketName   string `json:"bucketName"`
	Key          string `json:"key"`
	OutputBucket string `json:"outputBucket"`
}

func getData(w http.ResponseWriter, r *http.Request) {

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

	baseURL1 := "http://192.168.178.250:8085/getObject"

	params := url.Values{}
	params.Add("bucketName", cc.BucketName)
	params.Add("key", cc.Key)
	params.Add("formatJson", "false")

	parsedURL, err := url.Parse(baseURL1)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
	}

	// Attach the encoded query parameters to the URL
	parsedURL.RawQuery = params.Encode()

	// The fully constructed URL
	res1, err := spinhttp.Get(parsedURL.String())

	resp, err := ioutil.ReadAll(res1.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var lines []string
	if err := json.Unmarshal(resp, &lines); err != nil {
		panic(err)
	}

	customers := make(map[string]int)

	for _, line := range lines {
		lineParts := strings.Split(line, ",")
		if len(lineParts) == 8 {
			currCustomer := lineParts[6]
			if currCustomer == "" {
				continue
			}
			customers[currCustomer]++
		}
	}

	for key, value := range customers {
		assembledKey := "/key/" + key + "/" + randomString(10)
		valueJSON, err := json.Marshal(map[string]int{"value": value})
		if err != nil {
			panic(err)
		}

		baseURL := "http://192.168.178.250:8085/putObject"

		params = url.Values{}
		params.Add("bucketName", cc.OutputBucket)
		params.Add("key", assembledKey)

		parsedURL, err = url.Parse(baseURL)
		if err != nil {
			fmt.Println("Error parsing URL:", err)
		}

		// Attach the encoded query parameters to the URL
		parsedURL.RawQuery = params.Encode()

		// The fully constructed URL
		fullURL := parsedURL.String()

		//TODO http to put data
		_, err = spinhttp.Post(fullURL, "application/json", bytes.NewReader(valueJSON))
		if err != nil {
			http.Error(w, "Error reading putting Object from input with the key : "+cc.Key, http.StatusInternalServerError)
			return
		}
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

func randomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	rand.Seed(time.Now().UnixNano())
	s := make([]rune, length)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
