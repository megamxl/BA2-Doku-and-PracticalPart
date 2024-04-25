package function

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Message struct {
	Success bool    `json:"success"`
	Payload Payload `json:"payload"`
	Metrics Metrics `json:"metrics"`
}

type Payload struct {
	Test   string `json:"test"`
	N      uint64 `json:"n"`
	Result []int  `json:"result"`
	Time   int    `json:"time"`
}

type Metrics struct {
	MachineId  string `json:"machineid"`
	InstanceId string `json:"instanceid"`
	Cpu        string `json:"cpu"`
	Mem        string `json:"mem"`
	Uptime     string `json:"uptime"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var n = 100

	f := fibonacci()
	start := time.Now()
	var result int
	for i := 0; i < n; i++ {
		result = f()
	}
	elapsed := time.Since(start)

	m := Message{
		Success: true,
		Payload: Payload{
			Test: "matrix test",
			N:    uint64(result),
			Time: int(elapsed / time.Millisecond),
		},
		Metrics: Metrics{
			MachineId:  "",
			InstanceId: "instanceId",
			Cpu:        "cpuinfo",
			Mem:        "meminfo",
			Uptime:     "uptime",
		},
	}

	js, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(string(js))

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func fibonacci() func() int {
	before, val := 0, 1
	return func() int {
		ret := before
		before, val = val, before+val
		return ret
	}
}
