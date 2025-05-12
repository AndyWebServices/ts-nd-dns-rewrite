package main

import (
	"encoding/json"
	"log"
	"os"
)

// PostureObj Tailscale custom URL posture object
type PostureObj struct {
	Hostname      string   `json:"hostname"`
	UrlAttributes []string `json:"extra_hostnames"`
}

func (p PostureObj) generateRewrites() []NextdnsRewrite {
	rewrites := make([]NextdnsRewrite, len(p.UrlAttributes))
	for _, extraHostname := range p.UrlAttributes {
		rewrites = append(rewrites, NextdnsRewrite{
			Name:    p.Hostname + "." + urlAndyWebServicesCom,
			Content: extraHostname,
		})
	}
	return rewrites
}

func getLocalRewrites(fname string) []NextdnsRewrite {
	// Read YAML file and get list of possible objects
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	// Decode into slice of PostureObj
	var postureObjs []PostureObj
	err = json.Unmarshal(data, &postureObjs)
	if err != nil {
		log.Fatal(err)
	}

	// Total number of rewrites
	totalRewrites := 0
	for _, postureObj := range postureObjs {
		totalRewrites += len(postureObj.UrlAttributes)
	}

	// Gather all rewrites specified by the data file
	rewrites := make([]NextdnsRewrite, totalRewrites)
	for _, postureObj := range postureObjs {
		rewrites = append(rewrites, postureObj.generateRewrites()...)
	}
	return rewrites
}
