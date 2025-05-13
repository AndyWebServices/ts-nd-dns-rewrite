package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
)

const nextdnsTag = "tag:nextdns"
const urlAndyWebServicesCom = "andywebservices.com"
const tsHost = "api.tailscale.com"
const nextdnsHost = "api.nextdns.io"

const tsOrgName = "andywebservices.org.github"
const nextdnsProfileId = "8d2963"

type Config struct {
	TsApiToken      string
	NextDnsApiToken string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Grab TS_API_TOKEN from env
	tsApiToken := os.Getenv("TS_API_TOKEN")
	if tsApiToken == "" {
		log.Fatal("TS_API_TOKEN environment variable must be set")
	}

	// Grab NEXTDNS_API_TOKEN from env
	nextdnsApiToken := os.Getenv("NEXTDNS_API_TOKEN")
	if nextdnsApiToken == "" {
		log.Fatal("NEXTDNS_API_TOKEN environment variable must be set")
	}

	config := Config{
		TsApiToken:      tsApiToken,
		NextDnsApiToken: nextdnsApiToken,
	}

	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "App is a CLI example",
	}

	var cmdReload = &cobra.Command{
		Use:   "reload",
		Short: "Reload the app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("reload called")
			reload(config)
		},
	}

	var cmdPosture = &cobra.Command{
		Use:   "posture",
		Short: "Set posture for tailscale devices",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("posture called")
		},
	}

	rootCmd.AddCommand(cmdReload, cmdPosture)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
