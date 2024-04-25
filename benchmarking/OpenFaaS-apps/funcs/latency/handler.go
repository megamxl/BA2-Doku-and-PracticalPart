package function

import (
	"encoding/json"
	"log"
	"net/http"
)

type Message struct {
	Success bool    `json:"success"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	Test string `json:"test"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
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
}
