package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type RouteResponse struct {
	Routes []struct {
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance"`
	Duration  float64 `json:"duration"`
}

type RouteDetails struct {
	Coordinates [][]float64 `json:"coordinates"`
	Duration    float64     `json:"duration"`
	Distance    float64     `json:"distance"`
}

type GeoLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func main() {
	// Get start and end locations from user input
	// startLocation := "Dhaka"
	// endLocation := "Sylhet"

	var startLocation string
	var endLocation string

	fmt.Println("Enter your Start Location:")
	if _, err := fmt.Scanln(&startLocation); err != nil {
		log.Fatalf("Error reading start location: %v", err)
	}

	fmt.Println("Enter your End Location:")
	if _, err := fmt.Scanln(&endLocation); err != nil {
		log.Fatalf("Error reading end location: %v", err)
	}

	fmt.Printf("Start Location: %s\n", startLocation)
	fmt.Printf("End Location: %s\n", endLocation)

	// Get coordinates for start location
	startLat, startLon, err := getCoordinates(startLocation)
	if err != nil {
		log.Fatalf("Error getting coordinates for start location: %v", err)
	}

	fmt.Printf("startLat: %v, startLon: %v\n", startLat, startLon)

	// Get coordinates for end location
	endLat, endLon, err := getCoordinates(endLocation)
	if err != nil {
		log.Fatalf("Error getting coordinates for end location: %v", err)
	}

	fmt.Printf("endLat: %v, endLon: %v\n", endLat, endLon)

	// WebSocket server address
	serverAddr := "ws://localhost:8080/ws"

	// Get detailed route information from the API
	routeDetails, err := getRouteDetails(startLat, startLon, endLat, endLon)
	if err != nil {
		log.Fatalf("Error getting route details: %v", err)
	}

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

	// Function to continuously send location updates along the road
	go func() {
		for _, coord := range routeDetails.Coordinates {
			loc := Location{
				Latitude:  coord[1],
				Longitude: coord[0],
				Distance:  routeDetails.Distance,
				Duration:  routeDetails.Duration,
			}
			// Send location data to the server
			err := conn.WriteJSON(loc)
			if err != nil {
				log.Printf("Error sending location data: %v", err)
				continue
			}

			fmt.Printf("Location data sent: %+v\n", loc)

			// Wait for a short time before sending the next update
			time.Sleep(2 * time.Second) // Adjust the time interval as needed
		}

		fmt.Println("Reached the destination. Closing connection...")
		// Close the WebSocket connection gracefully after reaching the destination
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}
	}()

	// Wait for the Ctrl+C signal
	<-interrupt
	fmt.Println("Client stopped.")
}

func getRouteDetails(startLat, startLon, endLat, endLon float64) (RouteDetails, error) {
	start := fmt.Sprintf("%.6f,%.6f", startLon, startLat)
	end := fmt.Sprintf("%.6f,%.6f", endLon, endLat)

	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s;%s?geometries=geojson", start, end)

	resp, err := http.Get(url)
	if err != nil {
		return RouteDetails{}, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RouteDetails{}, fmt.Errorf("error reading response body: %v", err)
	}

	// Print response body for debugging
	fmt.Println("Response body:", string(body))

	var routeResponse RouteResponse
	if err := json.Unmarshal(body, &routeResponse); err != nil {
		return RouteDetails{}, fmt.Errorf("error decoding JSON response: %v", err)
	}

	// Extract coordinates from the response geometry
	var coordinates [][]float64
	if len(routeResponse.Routes) > 0 {
		geometry := routeResponse.Routes[0].Geometry
		coordinates = geometry.Coordinates
	}

	// Convert distance to kilometers
	distanceKm := routeResponse.Routes[0].Distance / 1000

	// Convert duration to minutes
	durationMinutes := routeResponse.Routes[0].Duration / 60

	routeDetails := RouteDetails{
		Coordinates: coordinates,
		Distance:    distanceKm,
		Duration:    durationMinutes,
	}

	return routeDetails, nil
}

func getCoordinates(location string) (float64, float64, error) {
	// Replace 'YOUR_API_KEY' with your actual LocationIQ API key
	apiKey := "pk.28216549993c88275e6b3d74005bfccf"

	// Query LocationIQ Geocoding API
	url := fmt.Sprintf("https://us1.locationiq.com/v1/search.php?key=%s&q=%s&format=json", apiKey, url.QueryEscape(location))
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("error querying LocationIQ Geocoding API: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("error reading LocationIQ Geocoding API response body: %v", err)
	}

	// Print response body for debugging
	fmt.Println("LocationIQ Geocoding API Response Body:", string(body))

	// Parse JSON response
	var locations []map[string]interface{}
	if err := json.Unmarshal(body, &locations); err != nil {
		return 0, 0, fmt.Errorf("error decoding LocationIQ Geocoding API JSON response: %v", err)
	}

	// Extract coordinates from the first location (assuming it's the most relevant)
	if len(locations) > 0 {
		lat, _ := strconv.ParseFloat(fmt.Sprintf("%v", locations[0]["lat"]), 64)
		lon, _ := strconv.ParseFloat(fmt.Sprintf("%v", locations[0]["lon"]), 64)
		return lat, lon, nil
	}

	return 0, 0, fmt.Errorf("no coordinates found for location: %s", location)
}

// fmt.Println("lat: ", lat)
// fmt.Println("lon: ", lon)

/*
package main

import (
	"fmt"
	"log"
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

	// Define coordinates representing points along the road from Teknaf to Tetulia
	roadCoordinates := []Location{
		{Latitude: 20.858170, Longitude: 92.288933}, // Teknaf
		{Latitude: 21.050603, Longitude: 92.233954},
		{Latitude: 21.153153, Longitude: 92.152828},
		{Latitude: 21.232102, Longitude: 92.161864},
		{Latitude: 21.318429, Longitude: 92.09321},
		{Latitude: 21.371041, Longitude: 92.120416},
		{Latitude: 21.525744, Longitude: 92.065231},
		{Latitude: 21.923604, Longitude: 92.058551},
		{Latitude: 22.054778, Longitude: 92.109246},
		{Latitude: 22.155017, Longitude: 92.071254},
		{Latitude: 22.279548, Longitude: 92.002652},
		{Latitude: 22.315935, Longitude: 91.923657},
		{Latitude: 22.291578, Longitude: 91.869872},
		{Latitude: 22.385049, Longitude: 91.762442},
		{Latitude: 22.944992, Longitude: 91.510014},
		{Latitude: 23.022898, Longitude: 91.366092},
		{Latitude: 23.321517, Longitude: 91.280983},
		{Latitude: 23.478584, Longitude: 91.132863},
		{Latitude: 23.530305, Longitude: 90.688534},
		{Latitude: 23.690369, Longitude: 90.546616},
		{Latitude: 23.731425, Longitude: 90.587845},
		{Latitude: 23.82816, Longitude: 90.563353},
		{Latitude: 23.919822, Longitude: 90.467527},
		{Latitude: 24.089123, Longitude: 90.194883},
		{Latitude: 24.139014, Longitude: 90.020612},
		{Latitude: 24.373868, Longitude: 89.876362},
		{Latitude: 24.419045, Longitude: 89.551604},
		{Latitude: 24.709824, Longitude: 89.395358},
		{Latitude: 24.942834, Longitude: 89.346546},
		{Latitude: 25.19655, Longitude: 89.388737},
		{Latitude: 25.44321, Longitude: 89.299912},
		{Latitude: 25.75219, Longitude: 89.253469},
		{Latitude: 25.811629, Longitude: 89.144322},
		{Latitude: 25.756587, Longitude: 88.673815},
		{Latitude: 25.86229, Longitude: 88.654341},
		{Latitude: 26.018396, Longitude: 88.468233},
		{Latitude: 26.199393, Longitude: 88.555023},
		{Latitude: 26.450938, Longitude: 88.565826},
		{Latitude: 26.482038, Longitude: 88.350381}, // Tetulia
	}

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

	// Function to continuously send location updates along the road
	go func() {
		for i, coord := range roadCoordinates {
			// Send location data to the server
			err := conn.WriteJSON(coord)
			if err != nil {
				log.Printf("Error sending location data: %v", err)
				continue
			}

			fmt.Printf("Location data sent: %+v\n", coord)

			// If it's not the last coordinate, wait for a short time before sending the next update
			if i < len(roadCoordinates)-1 {
				time.Sleep(5 * time.Second) // Adjust the time interval as needed
			}
		}

		fmt.Println("Reached the destination (Tetulia). Closing connection...")
		// Close the WebSocket connection gracefully after reaching the destination
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}
	}()

	// Wait for the Ctrl+C signal
	<-interrupt
	fmt.Println("Client stopped.")
}
*/
/*
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

// // Dhaka coordinates: latitude: 23.8103, longitude: 90.4125
// const (
//
//	DhakaLatitudeMin  = 23.6
//	DhakaLatitudeMax  = 24.0
//	DhakaLongitudeMin = 90.2
//	DhakaLongitudeMax = 90.6
//
// )

// Bangladesh coordinates: latitude: 20.34, longitude: 92.4125
const (
	BangladeshLatitudeMin  = 20.34 // Minimum latitude of Bangladesh
	BangladeshLatitudeMax  = 26.38 // Maximum latitude of Bangladesh
	BangladeshLongitudeMin = 88.01 // Minimum longitude of Bangladesh
	BangladeshLongitudeMax = 92.41 // Maximum longitude of Bangladesh
)

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
			case <-time.Tick(2 * time.Second): // Send location updates every 5 seconds
				// Generate random location data within Dhaka
				// location := Location{
				// 	Latitude:  rand.Float64()*(DhakaLatitudeMax-DhakaLatitudeMin) + DhakaLatitudeMin,
				// 	Longitude: rand.Float64()*(DhakaLongitudeMax-DhakaLongitudeMin) + DhakaLongitudeMin,
				// }

				location := Location{
					Latitude:  rand.Float64()*(BangladeshLatitudeMax-BangladeshLatitudeMin) + BangladeshLatitudeMin,
					Longitude: rand.Float64()*(BangladeshLongitudeMax-BangladeshLongitudeMin) + BangladeshLongitudeMin,
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

*/

/*

package main

import (
	"encoding/json"
	"fmt"
	"log"
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

// Define coordinates representing points along the road from Teknaf to Tetulia
var roadCoordinates = []Location{
	{Latitude: 20.858170, Longitude: 92.288933}, // Teknaf
	// Add more coordinates representing points along the road here
	{Latitude: 26.4820491, Longitude: 88.350376}, // Tetulia
}

func main() {
	// WebSocket server address
	serverAddr := "ws://localhost:8080/ws"

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

	// Function to continuously send location updates along the road
	go func() {
		for i, coord := range roadCoordinates {
			// Send location data to the server
			err := conn.WriteJSON(coord)
			if err != nil {
				log.Printf("Error sending location data: %v", err)
				continue
			}

			fmt.Printf("Location data sent: %+v\n", coord)

			// If it's not the last coordinate, wait for a short time before sending the next update
			if i < len(roadCoordinates)-1 {
				time.Sleep(5 * time.Second) // Adjust the time interval as needed
			}
		}

		fmt.Println("Reached the destination (Tetulia). Closing connection...")
		// Close the WebSocket connection gracefully after reaching the destination
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Error sending close message: %v", err)
		}
	}()

	// Wait for the Ctrl+C signal
	<-interrupt
	fmt.Println("Client stopped.")
}


*/

/*

Teknaf, Bangladesh

Latitude and longitude coordinates are: 20.858170, 92.288933.

Tetulia, Bangladesh

Latitude and longitude coordinates are: 26.4820491 Latitude

*/

/*
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
*/
