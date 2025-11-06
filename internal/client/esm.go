package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"time"

	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/checker"
)

// NAMClient handles API calls to NAM
type ESMClient struct {
	httpClient  *http.Client
	baseURL     string
	credentials ESMCredentials
	apiToken    string
}

type ESMCredentials struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

type ESMEntity struct {
	EntityType string        `json:"entity_type"`
	Properties ESMProperties `json:"properties"`
}

type ESMRequest struct {
	Entities  []ESMEntity `json:"entities"`
	Operation string      `json:"operation"`
}

type ESMProperties struct {
	RequestsOffering   string
	CreationSource     string
	RequestedByPerson  string
	RequestedForPerson string
	UserOptions        string
	DisplayLabel       string
	Description        string
	PublicScope        string
}

func NewESMClient(baseURL, username, password string) *ESMClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	return &ESMClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		credentials: ESMCredentials{
			Username: username,
			Password: password,
		},
	}
}

func (c *ESMClient) Authenticate() error {
	body, err := json.Marshal(c.credentials)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/auth/authentication-endpoint/authenticate/token", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")

	res, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	token, err := io.ReadAll(res.Body)

	if err != nil {
		return err
	}

	if res.StatusCode != 200 && res.StatusCode != 201 {
		return errors.New("smax login did not return OK")
	}
	c.apiToken = string(token)

	return nil
}

func (c *ESMClient) CreateRequest(result *checker.Result, vdcName, infra string) ESMRequest {

	preview := result.Output
	// lines := strings.Split(preview, "\n")

	// Format preview with HTML line breaks for ESM
	formattedPreview := strings.ReplaceAll(preview, "\n", "<br>")

	properties := ESMProperties{
		RequestsOffering:   "68905",
		CreationSource:     "CreationSourceEss",
		RequestedByPerson:  "200198",
		RequestedForPerson: "200198",
		UserOptions:        "{\"complexTypeProperties\":[{\"properties\":{\"DynamicComplexTypeRefName_c\":\"UserOption26bced6d955710349f8591eb08fe118dde2e_c\",\"changedUserOptionsForSimulation\":\"Person_c&\",\"history_Offering_User_Options_Price\":\"null\",\"Tjeneste_c\":\"73470\",\"Team_c\":\"67910\"}}]}",
		DisplayLabel:       fmt.Sprintf("VDC Infra Check - %s - %s", vdcName, infra),
		Description:        formattedPreview,
		PublicScope:        "Private",
	}

	request := ESMRequest{
		Entities: []ESMEntity{
			{
				EntityType: "Request",
				Properties: properties,
			},
		},
		Operation: "CREATE",
	}

	// bytes, _ := json.MarshalIndent(request, "  ", "")
	// fmt.Println("ESM Request Payload:", string(bytes))

	return request

}

func (c *ESMClient) SendRequest(request ESMRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	httpRequest, err := http.NewRequest("POST", fmt.Sprintf("%s/rest/%d/ems/bulk", c.baseURL, 938019087), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))

	res, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		return fmt.Errorf("smax request returned bad status code: %s", res.Body)
	}

	return nil
}
