package main

import (
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const port int = 42069

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

func writeResponse(w *response.Writer, code response.StatusCode, body []byte) {
	if err := w.WriteStatusLine(code); err != nil {
		log.Printf("error writing the status-line to the connection: %v\n", err)
	}
	if err := w.WriteHeaders(response.GetDefaultHeaders(len(body))); err != nil {
		log.Printf("error writing the headers to the connection: %v\n", err)
	}
	if _, err := w.Write(body); err != nil {
		log.Printf("error writing the body to the connection: %v\n", err)
	}

}

func Handler(w *response.Writer, r *request.Request) {

	switch r.RequestLine.RequestTarget {

	case "/yourproblem":
		writeResponse(w, response.StatusBadRequest, response400())
	case "/myproblem":
		writeResponse(w, response.StatusInternalServerError, response500())
	default:
		if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin/") {
			path := strings.TrimPrefix(r.RequestLine.RequestTarget, "/httpbin/")
			proxyHandler(w, path)
		} else {
			writeResponse(w, response.StatusOK, response200())
		}

	}

}

func proxyHandler(w *response.Writer, path string) {

	hdr := headers.NewHeaders()

	res, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		writeResponse(w, response.StatusInternalServerError, response500())
	} else {
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(response.GetDefaultHeaders(0))

		hdr.Delete("Content-Length")
		hdr.Set("Transfer-Encoding", "chunked")
		hdr.Replace("Content-Type", "text/plain")

		for {
			data := make([]byte, 30)
			n, err := res.Body.Read(data)
			if err != nil {
				break
			}
			w.WriteChunkedBody(data[:n])
		}
		w.WriteChunkedBodyDone()
		return
	}
}

func main() {

	server, err := server.Serve(port, Handler)
	if err != nil {
		log.Printf("error starting the server: %v\n", err)
	}

	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("error closing the server: %v\n", err)
		}
	}()

	log.Println("server started on port:", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	// server.Done()
	log.Println("server gracefully stopped")

}
