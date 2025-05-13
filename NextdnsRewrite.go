package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const profilesPath = "profiles"
const rewritesPath = "rewrites"

type NextdnsRewrite struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

func (r NextdnsRewrite) delete(config Config) {
	// Build device endpoint URL w/ id
	path, err := url.JoinPath(profilesPath, nextdnsProfileId, rewritesPath, r.Id)
	if err != nil {
		log.Fatal("Failed to join path:", err)
	}
	u := url.URL{
		Scheme: "https",
		Host:   nextdnsHost,
		Path:   path,
	}

	// Build DELETE request
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		log.Fatal("Error creating HTTP request:", err)
	}
	req.Header.Set("X-Api-Key", config.NextDnsApiToken)

	// Send DELETE request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		// If status code is 200 OK or 201 Created
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Success! Response: %s\n", body)
	} else {
		// Handle other status codes
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Request failed. Status code: %d %s\n", resp.StatusCode, body)
	}
}

func (r NextdnsRewrite) post(config Config) {
	nextdnsRewriteJson, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	// Build device endpoint URL
	path, err := url.JoinPath(profilesPath, nextdnsProfileId, rewritesPath)
	if err != nil {
		log.Fatal("Failed to join path:", err)
	}
	u := url.URL{
		Scheme: "https",
		Host:   nextdnsHost,
		Path:   path,
	}

	// Build DELETE request
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(nextdnsRewriteJson))
	if err != nil {
		log.Fatal("Error creating HTTP request:", err)
	}
	req.Header.Set("X-Api-Key", config.NextDnsApiToken)
	req.Header.Set("Content-Type", "application/json")

	// Send POST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()

	// Check the HTTP status code
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// If status code is 200 OK or 201 Created
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Success! Response: %s\n", body)
	} else {
		// Handle other status codes
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Request failed. Status code: %d %s\n", resp.StatusCode, body)
	}
}

func getNextdnsRewrites(config Config) []NextdnsRewrite {
	// Build GET url
	path, err := url.JoinPath(profilesPath, nextdnsProfileId, rewritesPath)
	if err != nil {
		log.Fatal("Failed to join path:", err)
	}
	u := url.URL{
		Scheme: "https",
		Host:   nextdnsHost,
		Path:   path,
	}

	// Build GET request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal("Error creating HTTP request:", err)
	}
	req.Header.Set("X-Api-Key", config.NextDnsApiToken)

	// Send GET request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()

	// Unmarshall data into slice
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading HTTP response:", err)
	}
	var data struct {
		NextdnsRewrites []NextdnsRewrite `json:"data"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Error parsing HTTP response:", err)
	}

	return data.NextdnsRewrites
}
