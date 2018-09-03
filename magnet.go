package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type magnet struct {
	Name string
	URL  string
}

func parseMagnet(line string) (m magnet) {
	if strings.Contains(line, "magnet:") {
		start := strings.Index(line, "magnet:")
		end := start + strings.Index(line[start:], `"`)

		m.URL = line[start:end]

		// Default name is the URL
		m.Name = m.URL

		// Try to find a "cleaner" name in the HTML
		match := ` title="`
		if strings.Contains(line, match) {
			start = len(match) + strings.Index(line, match)
			end = start + strings.Index(line[start:], `"`)
			m.Name = line[start:end]

			// Got one, return it
			return m
		}

		// If no title field in the HTML, attempt to parse from magnet link
		match = "dn="
		if strings.Contains(m.URL, match) {
			start = len(match) + strings.Index(m.URL, match)
			end = start + strings.Index(m.URL[start:], "&")
			if end == -1 {
				end = len(m.URL)
			}
			m.Name = m.URL[start:end]
		}
	}
	return m
}

func getMagnets(url string) (magnets []magnet, err error) {

	c := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := c.Get(url)
	if err != nil {
		return magnets, fmt.Errorf("could not get %s: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return magnets, fmt.Errorf("could not read page body: %v", err)
	}

	for _, line := range strings.Split(string(body), "\n") {
		if m := parseMagnet(line); m.URL != "" {
			magnets = append(magnets, m)
		}
	}

	return magnets, nil
}
