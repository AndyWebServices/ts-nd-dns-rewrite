package main

import (
	"fmt"
	"log"
	"sync"
)

func postMissingNextdnsRewrites(
	config Config,
	nextdnsRewriteByName map[string]NextdnsRewrite,
	tsRewriteByName map[string]NextdnsRewrite,
) {
	var wg sync.WaitGroup
	for rewriteName, rewrite := range tsRewriteByName {
		if _, ok := nextdnsRewriteByName[rewriteName]; !ok {
			fmt.Println(rewriteName, "->", rewrite.Content,
				"does not exist as nextdns nextdnsRewrite and will be created!")

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				rewrite.post(config)
			}(&wg)
		}
	}
	wg.Wait()
	fmt.Println("All new nextdns rewrites have been created!")
}

func deleteDeadNextdnsRewrites(
	config Config,
	nextdnsRewriteByName map[string]NextdnsRewrite,
	tsRewriteByName map[string]NextdnsRewrite,
) {
	// Delete nextdns entries that are missing or dead
	var wg sync.WaitGroup
	for rewriteName, nextdnsRewrite := range nextdnsRewriteByName {
		// Delete nextdnsRewrite no longer represented by a tailscale posture/device
		if _, ok := tsRewriteByName[rewriteName]; !ok {
			fmt.Println(rewriteName, "->", nextdnsRewrite.Content,
				"is not represented by any tailscale device/posture and will be deleted!")
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				nextdnsRewrite.delete(config)
			}(&wg)
		}
	}
	wg.Wait()
	fmt.Println("All dead nextdns rewrites have been deleted!")
}

func reload(config Config) {
	ch1 := make(chan map[string]TsDevice)
	ch2 := make(chan map[string]NextdnsRewrite)
	// Generate tsDeviceByName map
	go func() {
		// Grab tsDevices that contain nextdnsTag
		tsDevices := GetTsDevices(config, nextdnsTag, 10)
		tsDeviceByName := make(map[string]TsDevice, len(tsDevices))
		for _, tsDevice := range tsDevices {
			tsDeviceByName[tsDevice.Name] = tsDevice
		}
		ch1 <- tsDeviceByName
	}()
	// Generate nextdnsRewriteByName
	go func() {
		// Get all nextdnsRewrites from nextdns
		nextdnsRewrites := getNextdnsRewrites(config)
		nextdnsRewriteByName := make(map[string]NextdnsRewrite, len(nextdnsRewrites))
		for _, nextdnsRewrite := range nextdnsRewrites {
			nextdnsRewriteByName[nextdnsRewrite.Name] = nextdnsRewrite
		}
		ch2 <- nextdnsRewriteByName
	}()
	tsDeviceByName := <-ch1
	nextdnsRewriteByName := <-ch2

	// Merge tsRewrites into a big map. Check for collision detection and fail early if we see them
	tsRewriteByName := make(map[string]NextdnsRewrite, len(tsDeviceByName)*10)
	for _, tsDevice := range tsDeviceByName {
		for tsRewriteName, tsRewrite := range tsDevice.getRewriteByName() {
			if oldTsRewrite, ok := tsRewriteByName[tsRewriteName]; ok {
				log.Fatalf("rewrite %s -> %s already exists. Cannot be replaced with %s -> %s",
					oldTsRewrite.Name, oldTsRewrite.Content,
					tsRewriteName, tsRewrite.Content,
				)
			}
			tsRewriteByName[tsRewriteName] = tsRewrite
		}
	}

	// Create all rewrites represented by TS configs but missing in Nextdns
	var wg sync.WaitGroup
	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		postMissingNextdnsRewrites(config, nextdnsRewriteByName, tsRewriteByName)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		deleteDeadNextdnsRewrites(config, nextdnsRewriteByName, tsRewriteByName)
	}(&wg)
	wg.Wait()
}
