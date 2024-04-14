package router

import (
	"github.com/fasthttp/router"
	"github.com/nihankhan/locastream/internal/api"
)

const (
	home      = "/home"
	websocket = "/ws"
)

func Routers() *router.Router {
	r := router.New()

	r.GET(home, api.Home)
	r.GET(websocket, api.WebSocket)

	return r
}
