package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
)

var (
	tcpAddr  = "0.0.0.0:8081"
	httpAddr = "0.0.0.0:8080"
)

func getReqHeaders(r *http.Request) string {
	var headers string
	if reqHeadersBytes, err := json.Marshal(r.Header); err != nil {
		slog.Warn("Could not Marshal Req Headers")
	} else {
		headers = string(reqHeadersBytes)
	}
	return headers
}

func handleTCPconn(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Warn("Failed to close connection", "error", err.Error())
		}
	}()
	clientIP := conn.RemoteAddr().String()
	reader := bufio.NewReader(conn)
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				slog.Info("TCP connection closed by client", "source_ip", clientIP)
				return
			}
			// Log the error but don't exit - just close this connection
			slog.Warn("TCP read error", "source_ip", clientIP, "error", err.Error())
			return
		}
		_, err = conn.Write(bytes)
		if err != nil {
			slog.Warn("TCP write error", "source_ip", clientIP, "error", err.Error())
			return
		}
		slog.Info("Echoed data to TCP client", "source_ip", clientIP)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	headers := getReqHeaders(r)
	slog.Info("HTTP request received", "source_ip", clientIP, "method", r.Method, "path", r.URL.Path)
	w.Header().Set("Content-Type", "text/plain")

	hostname, err := os.Hostname()
	if err != nil {
		slog.Warn("Could not determine hostname")
	}

	if r.Method == http.MethodGet {
		_, err := w.Write([]byte("Hello from " + hostname + "\n" + "Request headers: " + headers + "\n"))
		if err != nil {
			slog.Warn("HTTP write error", "source_ip", clientIP, "error", err.Error())
			return
		}
		slog.Info("Returned home to client", "source_ip", clientIP)
	} else {
		if _, err := fmt.Fprintf(w, "Ignoring HTTP method %s\n", r.Method); err != nil {
			slog.Warn("Failed to write HTTP response", "source_ip", clientIP, "error", err.Error())
		}
		slog.Info("Served HTTP info page", "source_ip", clientIP)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	slog.Info("HTTP request received", "source_ip", clientIP, "method", r.Method, "path", r.URL.Path)

	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Warn("Error reading request body", "source_ip", clientIP, "error", err.Error())
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, err = w.Write(body)
		if err != nil {
			slog.Warn("HTTP write error", "source_ip", clientIP, "error", err.Error())
			return
		}
		slog.Info("Echoed data to HTTP client", "source_ip", clientIP, "bytes", len(body))
	} else {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprintf(w, "Echo server is running. Send POST request to echo content.\n"); err != nil {
			slog.Warn("Failed to write HTTP response", "source_ip", clientIP, "error", err.Error())
		}
		slog.Info("Served HTTP info page", "source_ip", clientIP)
	}
}

func startTCPServer(wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		slog.Error("TCP server error", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			slog.Warn("Failed to close connection", "error", err.Error())
		}
	}()

	slog.Info("TCP server listening", "address", tcpAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Warn("TCP accept error", "error", err.Error())
			continue
		}
		clientIP := conn.RemoteAddr().String()
		slog.Info("Accepting TCP connection", "source_ip", clientIP)
		go handleTCPconn(conn)
	}
}

func startHTTPServer(wg *sync.WaitGroup) {
	defer wg.Done()

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/echo", echoHandler)

	slog.Info("HTTP server listening", "address", httpAddr)

	err := http.ListenAndServe(httpAddr, nil)
	if err != nil {
		slog.Error("HTTP server error", "error", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) >= 2 {
		httpAddr = os.Args[1]
	}
	if len(os.Args) >= 3 {
		tcpAddr = os.Args[2]
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go startTCPServer(&wg)

	wg.Add(1)
	go startHTTPServer(&wg)

	slog.Info("Server started with TCP and HTTP support")
	slog.Info("TCP echo server", "address", tcpAddr)
	slog.Info("HTTP echo server", "address", httpAddr)

	wg.Wait()
}
