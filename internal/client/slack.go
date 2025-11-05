package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/checker"
)

// SlackClient handles Slack notifications
type SlackClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewSlackClient creates a new Slack client
func NewSlackClient(webhookURL string) *SlackClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	return &SlackClient{
		webhookURL: webhookURL,
		httpClient: httpClient,
	}
}

// Send sends a formatted notification to Slack
func (c *SlackClient) Send(result *checker.Result) error {
	if c.webhookURL == "" {
		return nil // Slack webhook not configured, skip
	}

	if !result.HasMismatches {
		return nil // No mismatches, no notification needed
	}

	payload := c.buildPayload(result)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack API returned status %d", resp.StatusCode)
	}

	return nil
}

// buildPayload creates the Slack Block Kit payload
func (c *SlackClient) buildPayload(result *checker.Result) map[string]interface{} {
	// Truncate output if too long
	preview := result.Output
	lines := strings.Split(preview, "\n")
	if len(lines) > 50 {
		lines = lines[:50]
		preview = strings.Join(lines, "\n") + "\n(… truncated …)"
	}

	blocks := []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]interface{}{
				"type": "plain_text",
				"text": fmt.Sprintf("VLAN OG PREFIX RAPPORT FOR %s", strings.ToUpper(result.VDCName)),
			},
		},
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": "vlan og prefixer funnet har ikke korrekt 'infrastructure' eller navn satt i Netbox, og må korrigeres.",
			},
		},
		{
			"type": "divider",
		},
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": fmt.Sprintf("```%s```", preview),
			},
		},
	}

	return map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":  "#FF7900",
				"blocks": blocks,
			},
		},
	}
}
