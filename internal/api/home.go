package api

import (
	"fmt"

	"github.com/valyala/fasthttp"
)

func Home(ctx *fasthttp.RequestCtx) {
	fmt.Println("Hello World!")

	fmt.Fprintln(ctx, "Welcome to the Real-Time Location Streaming Server...")
}
