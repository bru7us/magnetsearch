package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type server struct {
	site     string
	synoHost string

	user string
	pass string
}

func main() {
	port := flag.Uint("port", 8080, "HTTP server will listen on this port")
	site := flag.String("prefix", "http://www.openoffice.org/distribution/p2p/magnet.html?", "URL prefix for search terms")
	dsmHost := flag.String("dsm-host", "192.168.1.2:5001", "host:port for Synology DSM API")
	user := flag.String("user", "mytv", "Username for connection to Synology DownloadStation")
	pass := flag.String("pass", "", "Password for connection to Synology DownloadStation (overrides env DS_PASS)")

	flag.Parse()

	if *pass == "" {
		*pass = os.Getenv("DS_PASS")
	}

	s := &server{
		site:     *site,
		synoHost: *dsmHost,
		user:     *user,
		pass:     *pass,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSearch)
	mux.HandleFunc("/add/", s.handleAdd)
	mux.HandleFunc("/assets/", handleAssets)

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: mux,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	log.Fatal(srv.ListenAndServe())
}

func (s *server) handleAdd(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	m := q.Get("magnet")
	if m == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	q.Del("magnet")

	err := synoAddMagnet(s.synoHost, m+"&"+q.Encode(), s.user, s.pass)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var magnets []magnet

	// get query from request
	query := r.URL.Query().Get("q")

	resp := new(strings.Builder)

	resp.WriteString(`<html>
	<head>
		<title>Magnet Search</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
		<link href="assets/magnet.css" rel="stylesheet" type="text/css" />
		<script type="text/javascript" src="assets/magnet.js"></script>
	</head>
	<body>
		<div><b>Torrent magnets available from %s%s:</b></div><br />`)

	magnets, err := getMagnets(s.site + query)
	if err != nil {
		log.Printf("failed to get magnets: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for _, m := range magnets {
		resp.WriteString(fmt.Sprintf(`<a href="" onClick="if(confirm('Download %s?'))sendRequest('add/?magnet=%s'); return false;">%s</a><br />`, m.Name, m.URL, m.Name))
	}

	resp.WriteString(`</body></html>`)
	w.Write([]byte(fmt.Sprintf(resp.String(), s.site, query)))
}

func handleAssets(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/assets/")
	a, ok := assets[path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", a.contentType)
	w.Write(a.content)
}
