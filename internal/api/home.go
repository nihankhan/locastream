package api

import (
	"fmt"
	"html/template"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var HomeTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Real-Time Location Streaming</title>
    <link rel="stylesheet" href="https://unpkg.com/leaflet/dist/leaflet.css">
    <style>
        #map { height: 80vh; }
        .info { margin-top: 10px; }
    </style>
</head>
<body>
    <h1>Real-Time Location Streaming</h1>
    <div id="map"></div>
    <div class="info">
        <p>Duration: <span id="duration"></span></p>
        <p>Distance: <span id="distance"></span></p>
    </div>
    <script src="https://unpkg.com/leaflet/dist/leaflet.js"></script>
    <script>
        var map = L.map('map').setView([0, 0], 13);
        var markers = {}; // Object to store markers for each user

        L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
        }).addTo(map);

        var ws = new WebSocket("ws://" + window.location.host + "/ws");
        ws.onmessage = function(event) {
            var location = JSON.parse(event.data);
            var userId = location.userId; // Assuming each location object has a userId property

            // Check if a marker exists for the user, if not, create one
            if (!markers[userId]) {
                markers[userId] = L.marker([location.latitude, location.longitude]).addTo(map);
            } else {
                // If marker exists, update its position
                markers[userId].setLatLng([location.latitude, location.longitude]).update();
            }

            // Update duration and distance
            document.getElementById('duration').textContent = location.duration + " minutes";
            document.getElementById('distance').textContent = location.distance + " km";
        };
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
