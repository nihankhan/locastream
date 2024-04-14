package api

import (
	"fmt"
	"log"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebSocket(ctx *fasthttp.RequestCtx) {
	// Upgrade the connection to WebSocket
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Fatal(err)
				break
			}

			fmt.Println("Received Message: ", msg)

			err = conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Fatal(err)
				break
			}
		}
	})
	if err != nil {
		// If the upgrade fails, return an error response
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusInternalServerError)
		return
	}
}
