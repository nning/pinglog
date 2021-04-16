package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/go-ping/ping"
)

type response struct {
	Seq  int
	Sent time.Time
	Rtt  time.Duration
	Recv bool
	Dst  string
}

func (r response) String() string {
	return fmt.Sprintf("(seq:%d/rtt:%v/recv:%t)", r.Seq, r.Rtt, r.Recv)
}

func writeResponses() {
	x, _ := json.Marshal(&responses)
	err := ioutil.WriteFile(*output, x, 0644)
	if err != nil {
		panic(err)
	}
}

var destination = flag.String("d", "1.1.1.1", "destination")
var output = flag.String("o", "pinglog.json", "output path")
var interval = flag.Int("i", 10, "output write interval in seconds")
var count = flag.Int("c", 0, "count of packets to be sent")

var responses []response

func main() {
	flag.Parse()

	pinger, err := ping.NewPinger(*destination)
	if err != nil {
		panic(err)
	}

	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Println(responses[len(responses)-1].Recv)

		if !responses[len(responses)-1].Recv {
			fmt.Println("Warning: Got Ctrl+C before response to last ping, removed last log entry")
			responses = responses[:len(responses)-1]
		}

		writeResponses()

		fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
	}

	pinger.OnSend = func(packet *ping.Packet) {
		var r response
		r.Seq = packet.Seq
		r.Sent = time.Now()
		r.Dst = packet.Addr

		responses = append(responses, r)
	}

	pinger.OnRecv = func(packet *ping.Packet) {
		r := &responses[packet.Seq]

		r.Recv = true
		r.Rtt = packet.Rtt
	}

	if *count != 0 {
		pinger.Count = *count
	}

	// Listen for Ctrl-C.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pinger.Stop()
		}
	}()

	go func() {
		for {
			time.Sleep(time.Duration(*interval) * time.Second)
			writeResponses()
		}
	}()

	err = pinger.Run() // Blocks until finished.
	if err != nil {
		panic(err)
	}
}
