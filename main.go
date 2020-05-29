package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	var addr, path string
	flag.StringVar(&addr, "addr", "localhost:8080", "http service address")
	flag.StringVar(&path, "path", "/data/post", "http service address")

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: addr, Path: path}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	var data SolidData

	data.Velocity = 20
	data.SpecialError = "Error: No Error"
	data.ErrorStatus = 0
	data.EngineTemp = 26.4
	data.BatteryTemp = 28.2
	data.BatteryPercent = 87.6
	data.AverageVoltage = 3.7645
	data.AverageAmp = 12.87

	sender := time.NewTicker(time.Millisecond * 300)
	modifier := time.NewTicker(time.Millisecond * 200)
	defer sender.Stop()
	defer modifier.Stop()
	for {
		select {
		case <-done:
			return
		case <-sender.C:
			b, err := json.Marshal(data)
			if err != nil {
				log.Println(err)
				return
			}
			err = c.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-modifier.C:
			data.Velocity += 0.07
			data.EngineTemp += 0.09
			data.BatteryTemp += 0.07
			data.BatteryPercent -= 0.11
			data.AverageVoltage -= 0.09
			data.AverageAmp += 0.06
		case <-interrupt:
			log.Println("interrupt")

			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
