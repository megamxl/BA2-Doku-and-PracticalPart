package main

import (
	"encoding/json"
	"log"
	"net/http"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type Message struct {
	Success bool    `json:"success"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	Test string `json:"test"`
}

func main() {}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		//TODO overrided n to passed Value 2688834647444046
		m := Message{
			Success: true,
			Payload: Payload{
				Test: "latency test",
			},
		}

		js, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}
