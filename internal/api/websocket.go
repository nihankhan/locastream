package api

import (
	"fmt"
	"log"
	"sync"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

// Define a mutex to safely access the connections slice from multiple goroutines
var connectionsMutex sync.Mutex

// Slice to hold all WebSocket connections
var connections []*websocket.Conn

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
		// Allow all origins
		return true
	},
}

func WebSocket(ctx *fasthttp.RequestCtx) {
	// Upgrade the connection to WebSocket
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket read error:", err)
				break
			}

			fmt.Println("Received Message: ", string(msg))

			err = conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("WebSocket write error:", err)
				break
			}
		}
	})
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusInternalServerError)
	}
}

// BroadcastMessage broadcasts the message to all connected clients
func BroadcastMessage(msg []byte) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	// Iterate over all connected clients and send the message
	for _, conn := range connections {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			// Handle write error (e.g., connection closed)
			fmt.Println("Error writing message:", err)
		}
	}
}

// AddConnection adds a new WebSocket connection to the list of connections
func AddConnection(conn *websocket.Conn) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	connections = append(connections, conn)
}

// RemoveConnection removes a WebSocket connection from the list of connections
func RemoveConnection(conn *websocket.Conn) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()

	// Find and remove the connection from the slice
	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
}
