package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	authAPIName = "SYNO.API.Auth"
	dsAPIName   = "SYNO.DownloadStation.Task"
)

type infoAPIResponse struct {
	Data map[string]struct {
		Path string `json:"path"`
	} `json:"data"`
}

type authAPIResponse struct {
	Data struct {
		SID string `json:"sid"`
	} `json:"data"`
}

func synoAddMagnet(host, link, user, pass string) error {
	apiHost := "https://" + host + "/webapi"
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	c := http.Client{Transport: tr}
	resp, err := c.Get(apiHost + "/query.cgi?api=SYNO.API.Info&version=1&method=query&query=SYNO.API.Auth,SYNO.DownloadStation.Task")
	if err != nil {
		return fmt.Errorf("could not get API info: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read info response body: %v", err)
	}
	var info infoAPIResponse
	err = json.Unmarshal(body, &info)
	if err != nil {
		return fmt.Errorf("could not parse info body: %v", err)
	}

	if _, ok := info.Data[authAPIName]; !ok {
		return fmt.Errorf("could not find Auth API in API info response")
	}
	if _, ok := info.Data[dsAPIName]; !ok {
		return fmt.Errorf("could not find DownloadStation API in API info response")
	}

	resp, err = c.Get(fmt.Sprintf("%s/%v?api=%s&version=2&method=login&account=%s&passwd=%s&session=DownloadStation", apiHost, info.Data[authAPIName].Path, authAPIName, user, pass))
	if err != nil {
		return fmt.Errorf("could not authenticate: %v", err)
	}
	defer resp.Body.Close()

	// Defer logging out of the session
	defer c.Get(fmt.Sprintf("%s/%v?api=%s&version=2&method=logout&session=DownloadStation", apiHost, info.Data[authAPIName].Path, authAPIName))

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read auth response body: %v", err)
	}

	var auth authAPIResponse
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return fmt.Errorf("could not parse auth response body: %v", err)
	}

	if auth.Data.SID == "" {
		return fmt.Errorf("auth sid was empty")
	}

	v := make(url.Values)
	v.Add("api", dsAPIName)
	v.Add("version", "1")
	v.Add("method", "create")
	v.Add("uri", link)
	v.Add("_sid", auth.Data.SID)
	resp, err = c.PostForm(fmt.Sprintf("%s/%v", apiHost, info.Data[dsAPIName].Path), v)
	if err != nil {
		return fmt.Errorf("could not POST to add magnet: %v", err)
	}

	return nil
}
