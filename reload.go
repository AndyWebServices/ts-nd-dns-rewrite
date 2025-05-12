package main

import (
	"fmt"
	"log"
)

func reload(config Config) {
	// Grab tsDevices that contain nextdnsTag
	tsDevices := GetTsDevices(config, nextdnsTag)
	if len(tsDevices) == 0 {
		fmt.Println("No tsDevices found with tag " + nextdnsTag)
		return
	}

	// Get all nextdnsRewrites from nextdns
	nextdnsRewrites := getNextdnsRewrites(config)

	// Create maps for fast lookup
	tsDevicesByName := make(map[string]TsDevice, 100)
	for _, tsDevice := range tsDevices {
		tsDevicesByName[tsDevice.Name] = tsDevice
	}
	nextdnsRewritesByName := make(map[string]NextdnsRewrite, len(nextdnsRewrites))
	for _, nextdnsRewrite := range nextdnsRewrites {
		nextdnsRewritesByName[nextdnsRewrite.Name] = nextdnsRewrite
	}

	// Create all rewrites represented by TS configs but missing in Nextdns
	tsRewritesByName := make(map[string]NextdnsRewrite, len(tsDevicesByName))
	for _, tsDevice := range tsDevicesByName {
		rewrites := tsDevice.getRewrites()
		for _, rewrite := range rewrites {
			// Duplicate rewrites can occur if bad custom url postures are set
			if existingRewrite, ok := tsRewritesByName[rewrite.Name]; ok {
				log.Fatal("Duplicate rewrite names found:",
					existingRewrite.Name, "->", rewrite.Content,
					rewrite.Name, "->", rewrite.Content)
			}
			tsRewritesByName[rewrite.Name] = rewrite
			if _, ok := nextdnsRewritesByName[rewrite.Name]; !ok {
				fmt.Println(rewrite.Name, "->", rewrite.Content,
					"does not exist as nextdns nextdnsRewrite and will be created!")
				rewrite.post(config)
			}
		}
	}
	fmt.Println("All new nextdns rewrites have been created!")

	// Delete nextdns entries that are missing or dead
	for rewriteName, nextdnsRewrite := range nextdnsRewritesByName {
		// Delete dead nextdnsRewrite, aka content points to invalid ts device
		if _, ok := tsDevicesByName[nextdnsRewrite.Content]; !ok {
			fmt.Println(nextdnsRewrite.Content, "does not exist as tailscale host and will be deleted!")
			nextdnsRewrite.delete(config)
			continue
		}

		// Delete nextdnsRewrite no longer represented by a tailscale posture/device
		if _, ok := tsRewritesByName[rewriteName]; !ok {
			fmt.Println(rewriteName, "->", nextdnsRewrite.Content,
				"is not represented by any tailscale device/posture and will be deleted!")
			nextdnsRewrite.delete(config)
			continue
		}
	}
	fmt.Println("All dead nextdns rewrites have been deleted!")
}
