package main

import (
	"fmt"
	"httpfromtcp/internal"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/viper"
)

func response400() []byte {
	return []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
}

func response500() []byte {
	return []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
}

func response200() []byte {
	return []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
}

func Handler(w *internal.ResponseWriter, r *internal.Request) {

	switch r.RequestLine.RequestTarget {

	case "/yourproblem":
		writeResponse(w, internal.StatusBadRequest, response400())
	case "/myproblem":
		writeResponse(w, internal.StatusInternalServerError, response500())
	default:
		if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			path := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
			proxyHandler(w, path)
		} else {
			writeResponse(w, internal.StatusOK, response200())
		}

	}

}

func writeResponse(w *internal.ResponseWriter, code internal.HTTPStatusCode, body []byte) {
	if err := w.WriteStatusLine(code); err != nil {
		log.Printf("error writing the status-line to the connection: %v\n", err)
	}
	if err := w.WriteHeaders(internal.GetDefaultHeaders(len(body))); err != nil {
		log.Printf("error writing the headers to the connection: %v\n", err)
	}
	if _, err := w.Write(body); err != nil {
		log.Printf("error writing the body to the connection: %v\n", err)
	}

}

func proxyHandler(w *internal.ResponseWriter, path string) {

	hdr := internal.NewHeaders()

	res, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		writeResponse(w, internal.StatusInternalServerError, response500())
	} else {
		if err := w.WriteStatusLine(internal.StatusOK); err != nil {
			log.Printf("error writing the status-line to the connection: %v\n", err)
		}

		if err := w.WriteHeaders(internal.GetDefaultHeaders(0)); err != nil {
			log.Printf("error writing the headers to the connection: %v\n", err)
		}

		hdr.Delete("Content-Length")
		hdr.Set("Transfer-Encoding", "chunked")
		hdr.Replace("Content-Type", "text/plain")

		for {
			data := make([]byte, 30)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}
			if _, err := w.WriteChunkedBody(data[:n]); err != nil {
				log.Printf("error writing the chunked body to the connection: %v\n", err)
			}
		}
		if _, err := w.WriteChunkedBodyDone(); err != nil {
			log.Printf("error writing the end of chunked body to the connection: %v\n", err)
		}

		return
	}
}

// readConfig reads the config and loads the data
// accessible by viper
func readConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func main() {
	readConfig()
	proto := viper.GetString("protocol")
	addr := viper.GetString("address")

	var srv *server.Server
	switch strings.ToLower(proto) {
	case "udp":
		srv = server.NewServer(server.WithUDP(), server.WithAddr(addr))
	case "tcp":
		srv = server.NewServer(server.WithAddr(addr))
	default:
		log.Fatal("You must provid a valid transport protocol. see --help")

	}

	if err := srv.Serve(Handler); err != nil {
		log.Printf("error starting the server: %v\n", err)
	}

	defer func() {
		if err := srv.Close(); err != nil {
			log.Printf("error closing the server: %v\n", err)
		}
	}()

	log.Println("server started on", srv.Address())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	// server.Done()
	log.Println("server gracefully stopped")

}
