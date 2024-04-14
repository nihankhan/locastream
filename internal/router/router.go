package router

import (
	"github.com/fasthttp/router"
	"github.com/nihankhan/locastream/internal/api"
)

const (
	home = "/home"
)

func Routers() *router.Router {
	r := router.New()

	r.GET(home, api.Home)

	return r
}
