// mockshiny stands in for `Rscript -e "shiny::runApp('.', port=N)"` in tests,
// so modules/appserver can be exercised without a real R/Shiny installation.
//
// Invocation mirrors modules/appserver/instance.go exactly: argv[1] is "-e"
// and argv[2] is the R expression string, from which the port is extracted.
// Behavior can be steered per test by dropping marker files into the process
// working directory (which Instance.Start sets to the app's source dir):
//
//   - crash_immediately        exit(1) before printing "Listening on"
//   - start_delay_ms  (int)    sleep this many ms before printing "Listening on"
//   - crash_after_ms  (int)    exit(1) this many ms after starting
//
// Non-websocket requests get a plaintext body identifying the port that
// served them (so tests can verify session stickiness/load balancing) and
// echo back the appservR-* headers the proxy is expected to forward.
// Websocket upgrade requests get a minimal RFC 6455 handshake and then block
// until the peer closes the connection, so proxy websocket-close/session
// cleanup behavior can be exercised too.
package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var portRe = regexp.MustCompile(`port=(\d+)`)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: mockshiny -e \"shiny::runApp('.', port=NNNN)\"")
		os.Exit(2)
	}
	m := portRe.FindStringSubmatch(os.Args[2])
	if m == nil {
		fmt.Fprintln(os.Stderr, "mockshiny: could not find port= in expression:", os.Args[2])
		os.Exit(2)
	}
	port := m[1]

	cwd, _ := os.Getwd()

	if _, err := os.Stat(filepath.Join(cwd, "crash_immediately")); err == nil {
		fmt.Fprintln(os.Stderr, "mockshiny: simulated crash before start")
		os.Exit(1)
	}

	if ms := readIntMarker(cwd, "start_delay_ms"); ms > 0 {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mockshiny: listen error:", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketUpgrade(r) {
			handleWebSocket(w, r)
			return
		}
		w.Header().Set("X-Mock-Port", port)
		fmt.Fprintf(w, "mock-shiny-port=%s path=%s username=%s displayedname=%s appname=%s\n",
			port, r.URL.Path,
			r.Header.Get("appservR-username"),
			r.Header.Get("appservR-displayedname"),
			r.Header.Get("appservR-appname"))
	})

	server := &http.Server{Handler: mux}

	if ms := readIntMarker(cwd, "crash_after_ms"); ms > 0 {
		time.AfterFunc(time.Duration(ms)*time.Millisecond, func() {
			os.Exit(1)
		})
	}

	// Printed only once the listener is up, matching real Shiny's behavior of
	// logging "Listening on http://host:port" right before it starts serving.
	fmt.Println("Listening on http://127.0.0.1:" + port)

	server.Serve(ln)
}

func readIntMarker(dir, name string) int {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return n
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

const wsMagicGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijack unsupported", http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		return
	}
	defer conn.Close()

	h := sha1.New()
	h.Write([]byte(key + wsMagicGUID))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))

	buf.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	buf.WriteString("Upgrade: websocket\r\n")
	buf.WriteString("Connection: Upgrade\r\n")
	buf.WriteString("Sec-WebSocket-Accept: " + accept + "\r\n\r\n")
	buf.Flush()

	// Hold the connection open (echoing nothing back) until the peer closes
	// it, so tests can drive proxy-side websocket-disconnect cleanup.
	io.Copy(io.Discard, buf.Reader)
}
