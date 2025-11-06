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

// NAMClient handles API calls to NAM
type NAMClient struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
}

// NewNAMClient creates a new NAM API client
func NewNAMClient(baseURL, apiToken string) *NAMClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &NAMClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		apiToken:   apiToken,
	}
}

// FetchVxLANs fetches VxLANs from NAM
func (c *NAMClient) FetchVxLANs() ([]models.NAMVxLAN, error) {
	url := fmt.Sprintf("%s/api/ipam/vxlans/?expand=1", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VxLANs from NAM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NAM API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response struct {
		Results []models.NAMVxLAN `json:"results"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse NAM VxLAN response: %w", err)
	}

	return response.Results, nil
}
