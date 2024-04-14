package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nihankhan/locastream/internal/router"

	"github.com/valyala/fasthttp"
)

type Server struct {
	fastHttpServer *fasthttp.Server
}

func NewServer(router fasthttp.RequestHandler) *Server {
	return &Server{
		fastHttpServer: &fasthttp.Server{
			Handler: router,
		},
	}
}

func (s *Server) Start(addr string) {
	fmt.Println("Real-Time Location Streamng server running at: ", addr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		if err := s.fastHttpServer.ListenAndServe(addr); err != nil {
			log.Fatal(err)
		}
	}()

	<-done

	fmt.Println("Shuting down server...")

	if err := s.fastHttpServer.Shutdown(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server gracefully stopped!")
}

func main() {
	r := router.Routers()

	server := NewServer(r.Handler)

	server.Start(":8080")
}
