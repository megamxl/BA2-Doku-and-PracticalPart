package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	spinhttp "github.com/fermyon/spin/sdk/go/v2/http"
)

type Message struct {
	Success bool    `json:"success"`
	Payload Payload `json:"payload"`
	Metrics Metrics `json:"metrics"`
}

type Payload struct {
	Test   string  `json:"test"`
	N      uint64  `json:"n"`
	Result [][]int `json:"result"`
	Time   int     `json:"time"`
}

type Metrics struct {
	MachineId  string `json:"machineid"`
	InstanceId string `json:"instanceid"`
	Cpu        string `json:"cpu"`
	Mem        string `json:"mem"`
	Uptime     string `json:"uptime"`
}

func main() {}

func init() {
	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {

		//TODO overrided n to passed Value 2688834647444046
		nString := r.URL.Query().Get("n")

		var n int

		if nString != "" {
			atoi, err := strconv.Atoi(nString)
			if err == nil {
				n = atoi
			} else {
				n = 100
			}
		}

		start := time.Now()
		res := matrix(n)
		elapsed := time.Since(start)

		m := Message{
			Success: true,
			Payload: Payload{
				Test:   "matrix test",
				N:      uint64(n),
				Result: res,
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

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})
}

func randomTable(n int) [][]int {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	m := make([][]int, n)
	for i := 0; i < n; i++ {
		m[i] = make([]int, n)
		for j := 0; j < n; j++ {
			m[i][j] = r1.Intn(100)
		}
	}
	return m
}

func matrix(n int) [][]int {
	matrixA := randomTable(n)
	matrixB := randomTable(n)
	matrixMult := make([][]int, n)

	for i := 0; i < len(matrixA); i++ {
		matrixMult[i] = make([]int, n)
		for j := 0; j < len(matrixB); j++ {
			sum := 0
			for k := 0; k < len(matrixA); k++ {
				sum += matrixA[i][k] * matrixB[k][j]
			}
			matrixMult[i][j] = sum
		}
	}

	return matrixMult
}
