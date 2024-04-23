package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
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

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		//TODO overrided n to passed Value 2688834647444046
		var n uint64 = 2688834647444046

		start := time.Now()
		var result []int = factors(n)
		elapsed := time.Since(start)

		m := Message{
			Success: true,
			Payload: Payload{
				Test:   "cpu test",
				N:      n,
				Result: result,
				Time:   int(elapsed / time.Millisecond),
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
	})
}

func main() {}

func factors(num uint64) []int {
	var n_factors = make([]int, 0)
	sqrtNum := uint64(math.Sqrt(float64(num))) // Cast the square root to uint64

	for i := uint64(1); i <= sqrtNum; i++ {
		if num%i == 0 {
			n_factors = append(n_factors, int(i)) // Cast i to int before appending
			if num/i != i {
				n_factors = append(n_factors, int(num/i)) // Cast num/i to int before appending
			}
		}
	}

	sort.Ints(n_factors)
	return n_factors
}
