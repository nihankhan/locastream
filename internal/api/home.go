package api

import (
	"fmt"
	"html/template"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

//<!-- HomeTemplate is the HTML template for the home page -->

var HomeTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Real-Time Location Streaming</title>
    <!-- Leaflet CSS -->
    <link rel="stylesheet" href="https://unpkg.com/leaflet/dist/leaflet.css" integrity="sha384-SO/mf3RVeZ2jnPExIzHnUV73zTfSxEhJEMHDs5gIVBUurvZI6U7E3lItPb2zFZB1" crossorigin="">

    <style>
        #map { height: 400px; }
    </style>
</head>
<body>
    <h1>Real-Time Location Streaming</h1>
    <!-- Map container -->
    <div id="map"></div>

    <!-- Leaflet.js -->
    <script src="https://unpkg.com/leaflet/dist/leaflet.js" integrity="sha384-KyZXEAg3QhqLMpG8r+Z/+CUSsFZOnLymxqM0i6A3gVmG4oqI2ansPPFjkFV5DI1b" crossorigin=""></script>

    <script>
        // Function to initialize Leaflet map and WebSocket connection
        function initializeMap() {
            // Initialize the map
            var map = L.map('map').setView([0, 0], 13);

            // Add a tile layer from OpenStreetMap
            L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
                attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
            }).addTo(map);

            // Initialize a marker for the user's location
            var marker = L.marker([0, 0]).addTo(map);

            // Establish WebSocket connection
            var ws = new WebSocket("ws://" + window.location.host + "/ws");

            ws.onopen = function(event) {
                console.log("WebSocket connection established.");
            };

            ws.onmessage = function(event) {
                console.log("Received message:", event.data);
                // Parse and process location data received from server
                var location = JSON.parse(event.data);
                updateLocation(location);
            };

            ws.onclose = function(event) {
                console.log("WebSocket connection closed.");
            };

            function updateLocation(location) {
                // Update marker position with the latest location data
                marker.setLatLng([location.latitude, location.longitude]).update();
            }
        }

        // Call the initializeMap function when the document is ready
        document.addEventListener('DOMContentLoaded', function() {
            initializeMap();
        });
    </script>
</body>
</html>
`

func Home(ctx *fasthttp.RequestCtx) {
	fmt.Println("Hello World!")

	// Check if the request is a WebSocket upgrade request
	if websocket.FastHTTPIsWebSocketUpgrade(ctx) {
		WebSocket(ctx)
		return
	}

	// Set the content type to text/html
	ctx.SetContentType("text/html")

	// Write the HTML template to the response body
	// fmt.Fprintf(ctx, HomeTemplate)

	// Serve the HTML template for the home page
	tmpl, err := template.New("home").Parse(HomeTemplate)
	if err != nil {
		log.Println("Error parsing template:", err)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}

	// Execute the template and write the response
	err = tmpl.Execute(ctx, nil)
	if err != nil {
		log.Println("Error executing template:", err)
		ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
		return
	}
}
