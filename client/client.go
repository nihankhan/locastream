package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func main() {
	// WebSocket server address
	serverAddr := "ws://localhost:8080/ws"

	// Initialize a random number generator
	rand.Seed(time.Now().UnixNano())

	// Handle Ctrl+C signal to gracefully close the connection
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Establish a WebSocket connection to the server
	u, _ := url.Parse(serverAddr)
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("Error connecting to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Function to continuously send location updates
	go func() {
		for {
			select {
			case <-time.Tick(5 * time.Second): // Send location updates every 5 seconds
				// Generate random location data for demonstration
				location := Location{
					Latitude:  rand.Float64()*180 - 90,
					Longitude: rand.Float64()*360 - 180,
				}

				// Marshal location data into JSON
				locationData, err := json.Marshal(location)
				if err != nil {
					log.Printf("Error marshaling location data: %v", err)
					continue
				}

				// Send location data to the server
				err = conn.WriteMessage(websocket.TextMessage, locationData)
				if err != nil {
					log.Printf("Error sending location data: %v", err)
					continue
				}

				fmt.Printf("Location data sent: %+v\n", location)
			case <-interrupt:
				fmt.Println("Interrupt signal received. Closing connection...")
				// Close the WebSocket connection gracefully
				err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Printf("Error sending close message: %v", err)
				}
				return
			}
		}
	}()

	// Wait for the Ctrl+C signal
	<-interrupt
	fmt.Println("Client stopped.")
}
