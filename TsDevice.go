package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"sync"
)

const devicePath = "device"
const attributesPath = "attributes"
const tailnetPath = "tailnet"
const devicesPath = "devices"

const tsApiV2Path = "/api/v2"

func tsGetRequest[T any](config Config, urlEndpoint string) T {
	// Build GET url
	u, err := url.Parse(urlEndpoint)
	if err != nil {
		log.Fatal("Error parsing URL:", err)
	}
	u.Scheme = "https"

	// Build GET req
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Fatal("Error creating HTTP request:", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.TsApiToken)

	// Send GET req
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatal("Error sending HTTP request. Status code:", resp.StatusCode, "Status:", resp.Status)
	}

	// Read body into
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading HTTP response:", err)
	}
	var data T
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal("Error parsing HTTP response:", err, "body:", string(body))
	}
	return data
}

type TsDevice struct {
	NodeId   string   `json:"nodeId"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	Hostname string   `json:"hostname"`

	Attributes     map[string]any `json:"attributes"`
	ExtraHostnames []string
}

func (d TsDevice) getAWSHostname() string {
	return fmt.Sprintf("%s.%s", url.PathEscape(d.Hostname), urlAndyWebServicesCom)
}

func (d TsDevice) setCustomUrls(config Config) {
	// Get device attributes
	path, err := url.JoinPath(tsApiV2Path, devicePath, d.NodeId, attributesPath)
	if err != nil {
		log.Fatal("Error joining path:", err)
	}
	u := url.URL{
		Scheme: "https",
		Host:   tsHost,
		Path:   path,
	}
	type AttributesStruct struct {
		Attributes map[string]any `json:"attributes"`
	}
	d.Attributes = tsGetRequest[AttributesStruct](config, u.String()).Attributes

	// Search for attributes that contain custom:url:{extraHostname}
	re := regexp.MustCompile(`custom:url:([a-zA-Z0-9.-]+)`)
	for attributeKey, attributeValue := range d.Attributes {
		matches := re.FindStringSubmatch(attributeKey)
		if len(matches) > 1 {
			extraHostname := matches[1]
			ok := attributeValue.(bool)
			if ok {
				d.ExtraHostnames = append(d.ExtraHostnames, extraHostname)
			}
		}
	}

	fmt.Printf("Device %s has extraHostnames -> %s\n", d.Hostname, d.ExtraHostnames)
}

func (d TsDevice) getRewriteByName() map[string]NextdnsRewrite {
	rewrites := make(map[string]NextdnsRewrite, 1+len(d.ExtraHostnames))

	nextdnsRewrite := NextdnsRewrite{Name: d.getAWSHostname(), Content: d.Name}
	rewrites[nextdnsRewrite.Name] = nextdnsRewrite
	for _, extraHostname := range d.ExtraHostnames {
		rewrites[extraHostname] = NextdnsRewrite{Name: extraHostname, Content: d.Name}
	}

	return rewrites
}

func GetTsDevices(config Config, tagName string, workerCount int) []TsDevice {
	path, err := url.JoinPath(tsApiV2Path, tailnetPath, tsOrgName, devicesPath)
	if err != nil {
		log.Fatal("Error joining path:", err)
	}
	u := url.URL{
		Scheme: "https",
		Host:   tsHost,
		Path:   path,
	}
	type DevicesStruct = struct {
		Devices []TsDevice `json:"devices"`
	}
	data := tsGetRequest[DevicesStruct](config, u.String())

	// Init sync primitives
	var wg sync.WaitGroup
	resultChan := make(chan TsDevice, len(data.Devices))
	jobChan := make(chan TsDevice, len(data.Devices))
	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tsDevice := range jobChan {
				if slices.Contains(tsDevice.Tags, tagName) {
					tsDevice.setCustomUrls(config)
					resultChan <- tsDevice
				}
			}
		}()
	}

	// Send jobs
	for _, tsDevice := range data.Devices {
		jobChan <- tsDevice
	}
	close(jobChan)

	// Go routine to close resultChan
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var tsDevices []TsDevice
	for tsDevice := range resultChan {
		tsDevices = append(tsDevices, tsDevice)
	}
	return tsDevices
}
