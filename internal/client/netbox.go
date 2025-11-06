package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/models"
)

// NetboxClient handles API calls to Netbox
type NetboxClient struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
}

// NewNetboxClient creates a new Netbox API client
func NewNetboxClient(baseURL, apiToken string) *NetboxClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &NetboxClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		apiToken:   apiToken,
	}
}

// FetchVLANs fetches VLANs from Netbox for a specific site
func (c *NetboxClient) FetchVLANs(siteID int) ([]models.NetboxVLAN, error) {
	url := fmt.Sprintf("%s/api/ipam/vlans/?site_id=%d&limit=1000", c.baseURL, siteID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VLANs from Netbox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Netbox API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Results []models.NetboxVLAN `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Netbox VLAN response: %w", err)
	}

	return response.Results, nil
}

// FetchPrefixes fetches prefixes from Netbox for a specific site
func (c *NetboxClient) FetchPrefixes(siteID int) ([]models.NetboxPrefix, error) {
	url := fmt.Sprintf("%s/api/ipam/prefixes/?site_id=%d&limit=1000", c.baseURL, siteID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch prefixes from Netbox: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Netbox API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Results []models.NetboxPrefix `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Netbox prefix response: %w", err)
	}

	return response.Results, nil
}
