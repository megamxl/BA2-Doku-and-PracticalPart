package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
	"github.com/fermyon/spin/sdk/go/v2/pg"
	reddis "github.com/fermyon/spin/sdk/go/v2/redis"
)

type Package struct {
	ID                 int    `json:"id"`
	TrackingNumber     string `json:"trackingNumber"`
	Sender             string `json:"sender"`
	Recipient          string `json:"recipient"`
	OriginAddress      string `json:"originAddress"`
	DestinationAddress string `json:"destinationAddress"`
	Weight             int    `json:"weight"`
	Status             string `json:"status"`
}

type PackageRequest struct {
	Sender             string `json:"sender"`
	Recipient          string `json:"recipient"`
	OriginAddress      string `json:"originAddress"`
	DestinationAddress string `json:"destinationAddress"`
	Weight             int64  `json:"weight"`
}

type ApiResponse struct {
	TrackingNumber string `json:"trackingNumber"`
}

type ChangePackageSate struct {
	TrackingNumber string `json:"trackingNumber"`
	Status         string `json:"status"`
}

func init() {
	rand.Seed(time.Now().UnixNano())

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		addr := "postgresql://exampleuser:examplepass@192.168.178.109:5432/exampledb"

		db := pg.Open(addr)
		defer db.Close()

		red := reddis.NewClient("redis://192.168.178.109:6379")

		if r.Method == "GET" {
			selectByID(*db, *red, w, r)
			return
		}

		if !checkBasicAuthCreds(w, r) {
			http.Error(w, "Not Authorized", http.StatusMethodNotAllowed)
			return
		}

		if r.Method == "PUT" {
			updateById(*db, *red, w, r)
			return
		}

		if r.Method == "POST" {
			createPackage(*db, *red, w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return

	})
}

func createPackage(db sql.DB, red reddis.Client, w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	var reqBody PackageRequest

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(body, &reqBody); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// addr is the environment variable set in `spin.toml` that points to the
	// address of the Mysql server.
	gen_trackingNumber := generateTrackingNumber()

	query := `INSERT INTO packages ( tracking_number, sender, recipient, origin_address, destination_address, weight, status) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = db.Exec(query, gen_trackingNumber, reqBody.Sender, reqBody.Recipient, reqBody.OriginAddress, reqBody.DestinationAddress, reqBody.Weight, "Data Received")
	if err != nil {
		http.Error(w, "Cant create Package due to SQL Problem", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"trackingNumber":     gen_trackingNumber,
		"sender":             reqBody.Sender,
		"recipient":          reqBody.Recipient,
		"originAddress":      reqBody.OriginAddress,
		"destinationAddress": reqBody.DestinationAddress,
		"weight":             reqBody.Weight,
		"status":             "Data Received",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshalling data: %v", err)
	}

	response := ApiResponse{
		TrackingNumber: gen_trackingNumber,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	err = red.Set(gen_trackingNumber, jsonData)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println((time.Now().UnixNano() - start.UnixNano()) / 1000000)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(220)
	w.Write(responseBytes) // Send the JSON response

}

func checkBasicAuthCreds(w http.ResponseWriter, r *http.Request) bool {
	username, password, ok := r.BasicAuth()
	if !ok {
		// If there's an error or no credentials provided, return an error
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	// Validate the credentials (this is just a dummy example)
	if username != "admin" || password != "admin" {
		// If the credentials are incorrect, return an error
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func generateTrackingNumber() string {
	// Using current Unix time to ensure a unique base
	timestamp := time.Now().Unix()

	// Generating a random number to append to the timestamp
	randomPart := rand.Intn(100000) // Random number between 0 and 99999

	// Formatting as a string, combining both elements
	trackingNumber := fmt.Sprintf("TN%d%05d", timestamp, randomPart)

	return trackingNumber
}

func selectByID(db sql.DB, red reddis.Client, w http.ResponseWriter, r *http.Request) {

	trackingNumber := r.URL.Query().Get("trackingNumber")
	if trackingNumber == "" {
		// If no trackingNumber is provided, return an error
		http.Error(w, "Tracking number is required", http.StatusBadRequest)
		return
	}

	bytes, err := red.Get(trackingNumber)

	if len(bytes) > 0 {

		var cahced Package

		err = json.Unmarshal(bytes, &cahced)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			w.Write(bytes)
			return
		}
	}

	// Define your SQL query
	query := `SELECT * FROM packages WHERE tracking_number = $1`

	// Execute the query
	row, err := db.Query(query, trackingNumber)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer row.Close()

	// Assuming you're reading one result, adjust as necessary for your application
	if !row.Next() {
		http.Error(w, "No package found with the given tracking number", http.StatusNotFound)
		return
	}

	//TODO make DTO and do not return the whole package
	var parcel Package
	if err := row.Scan(&parcel.ID, &parcel.TrackingNumber, &parcel.Sender, &parcel.Recipient, &parcel.OriginAddress, &parcel.DestinationAddress, &parcel.Weight, &parcel.Status); err != nil {
		http.Error(w, "Failed to read package data", http.StatusInternalServerError)
		return
	}

	// Send back the package information as JSON
	responseBytes, err := json.Marshal(parcel)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	red.Set(parcel.TrackingNumber, responseBytes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
	return
}

func updateById(db sql.DB, red reddis.Client, w http.ResponseWriter, r *http.Request) {

	var reqBody ChangePackageSate

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(body, &reqBody); err != nil {
		fmt.Println("hello")
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if !checkIfStateIsAllowed(reqBody.Status) {
		http.Error(w, "Not Supported Sate", http.StatusBadRequest)
	}

	query := "UPDATE packages SET status = $1 WHERE tracking_number = $2"

	ret, err := db.Exec(query, reqBody.Status, reqBody.TrackingNumber)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Cant update Package state due to SQL Problem", http.StatusInternalServerError)
		return
	}

	affected, err := ret.RowsAffected()
	if err != nil {
		http.Error(w, "Cant update Package state due to SQL Problem", http.StatusInternalServerError)
		return
	}

	if affected != 1 {
		http.Error(w, "No Package with this id is available ", http.StatusInternalServerError)
		return
	}

	query1 := `SELECT * FROM packages WHERE tracking_number = $1`

	// Execute the query
	row, err := db.Query(query1, reqBody.TrackingNumber)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		return
	}
	defer row.Close()

	// Assuming you're reading one result, adjust as necessary for your application
	if !row.Next() {
		http.Error(w, "No package found with the given tracking number", http.StatusNotFound)
		return
	}

	//TODO make DTO and do not return the whole package
	var parcel Package
	if err := row.Scan(&parcel.ID, &parcel.TrackingNumber, &parcel.Sender, &parcel.Recipient, &parcel.OriginAddress, &parcel.DestinationAddress, &parcel.Weight, &parcel.Status); err != nil {
		http.Error(w, "Failed to read package data", http.StatusInternalServerError)
		return
	}

	// Send back the package information as JSON
	responseBytes, err := json.Marshal(parcel)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	red.Set(reqBody.TrackingNumber, responseBytes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func checkIfStateIsAllowed(state string) bool {
	if state == "in-transit" || state == "in-nearest-facility" || state == "in-delivery" || state == "delivered" {
		return true
	}
	return false
}

func main() {}
