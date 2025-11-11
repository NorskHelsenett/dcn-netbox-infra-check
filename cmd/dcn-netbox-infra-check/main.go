package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/checker"
	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/client"
	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create API clients
	netboxClient := client.NewNetboxClient(cfg.NetboxURL, cfg.NetboxAPIToken)
	namClient := client.NewNAMClient(cfg.NAMURL, cfg.NAMAPIToken)

	// Create Slack client
	// slackClient := client.NewSlackClient(cfg.SlackWebhook)

	// Fetch NAM VxLANs once (shared across all checks)
	namVxLANs, err := namClient.FetchVxLANs()
	if err != nil {
		log.Fatalf("✗ Failed to fetch NAM VxLANs: %v", err)
	}

	if len(namVxLANs) == 0 {
		log.Fatal("✗ No NAM VxLANs fetched - check API URL or token")
	}

	// Process each check
	for _, check := range cfg.Checks {
		fmt.Printf("\n\n")
		fmt.Printf("==================================\n")
		fmt.Printf("Sjekker datasenter %s\n", strings.ToUpper(check.DCName))
		fmt.Printf("==================================\n\n")

		// Fetch Netbox data for this site
		netboxVLANs, err := netboxClient.FetchVLANs(check.NetboxSiteID)
		if err != nil {
			log.Fatalf("✗ Failed to fetch Netbox VLANs for site %d: %v", check.NetboxSiteID, err)
		}

		netboxPrefixes, err := netboxClient.FetchPrefixes(check.NetboxSiteID)
		if err != nil {
			log.Fatalf("✗ Failed to fetch Netbox Prefixes for site %d: %v", check.NetboxSiteID, err)
		}

		if len(netboxVLANs) == 0 {
			log.Fatalf("✗ No Netbox VLANs fetched for site %d - check API URL or token", check.NetboxSiteID)
		}

		// Perform checks
		result := checker.Check(
			check.DCName,
			check.Infra,
			netboxVLANs,
			netboxPrefixes,
			namVxLANs,
			cfg,
		)

		// Print results to console
		fmt.Print(result.Output)

		// Send to Slack if there are mismatches
		if result.HasMismatches {
			// if cfg.SlackWebhook != "" {
			// 	if err := slackClient.Send(result); err != nil {
			// 		log.Printf("✗ Failed to send Slack notification: %v", err)
			// 	}
			// }

			esmClient := client.NewESMClient(cfg.ESMURL, cfg.ESMUser, cfg.ESMPassword, cfg.ESMTenantID)

			err = esmClient.Authenticate()
			if err != nil {
				log.Fatalf("✗ Failed to authenticate to ESM: %v", err)
			}
			request := esmClient.CreateRequest(result, check.DCName, check.Infra, cfg)
			err = esmClient.SendRequest(request)
			if err != nil {
				log.Fatalf("✗ Failed to send ESM request: %v", err)
			}

		}
	}

	fmt.Printf("\n======================\n")
	fmt.Printf("Alle sjekker fullført!\n")
	fmt.Printf("======================\n\n")
}
