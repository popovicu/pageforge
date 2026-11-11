// devserver is a local HTTP file server for previewing a Pageforge site.
//
// It serves files from a directory using Go's standard http.FileServer, with
// caching disabled so a rebuilt site shows up on the next page load.
// Intended for use via `bazel run` with the site_serve rule, which
// handles tar extraction before launching this binary.
//
// Flags:
//
//	--dir      Directory to serve
//	--port     Port to listen on (default: 8080)
//	--host     Address to listen on (default: loopback only). Pass
//	           0.0.0.0 to reach the site from other devices.
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

func run(args []string) error {
	fs := flag.NewFlagSet("devserver", flag.ContinueOnError)
	dir := fs.String("dir", "", "Directory to serve")
	port := fs.Int("port", 8080, "Port to listen on")
	host := fs.String("host", "", "Address to listen on (default: loopback only; 0.0.0.0 exposes the site to the network)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *dir == "" {
		return fmt.Errorf("--dir is required")
	}

	// Loopback by default, on both families, so "localhost" answers
	// whichever address the client resolves first.
	hosts := []string{"127.0.0.1", "::1"}
	shown := "localhost"
	if *host != "" {
		hosts = []string{*host}
		shown = *host
	}

	var listeners []net.Listener
	for _, h := range hosts {
		ln, err := net.Listen("tcp", net.JoinHostPort(h, strconv.Itoa(*port)))
		if err != nil {
			if len(listeners) > 0 {
				continue // e.g. no IPv6 loopback on this machine
			}
			return err
		}
		listeners = append(listeners, ln)
	}

	fmt.Printf("Serving site at http://%s\n", net.JoinHostPort(shown, strconv.Itoa(*port)))
	srv := &http.Server{Handler: handler(*dir)}
	errc := make(chan error, len(listeners))
	for _, ln := range listeners {
		go func() { errc <- srv.Serve(ln) }()
	}
	return <-errc
}

// handler serves dir with caching turned off. Every file in a site tarball
// carries the same fixed mtime, so Last-Modified never changes between
// builds: browsers would treat old pages as fresh for years, or revalidate
// them to a 304, and keep showing a stale build.
func handler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("If-Modified-Since")
		r.Header.Del("If-None-Match")
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	})
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "devserver: %v\n", err)
		os.Exit(1)
	}
}
