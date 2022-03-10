package libs

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

type MonitorConnections struct {
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	ID         string `json:"id,omitempty"`
	Desc       string `json:"description,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	CreatedBy  string `json:"createdBy,omitempty"`
	ModifiedAt string `json:"modifiedAt,omitempty"`
	ModifiedBy string `json:"modifiedBy,omitempty"`
	URL        string `json:"url,omitempty"`
	UserName   string `json:"username,omitempty"`
}

func GiveConnectionIDS(token string) ([]MonitorConnections, error) {

	dep := os.Getenv(EnvKeySumoEnvironment)

	endpoint := fmt.Sprintf("https://api.%s.sumologic.com/api/v1/connections?limit=1000&token=%s", dep, token)
	if dep == "us1" {
		endpoint = fmt.Sprintf("https://api.sumologic.com/api/v1/connections?limit=1000&token=%s", token)
	}

	cl := resty.New()

	resp, err := cl.R().SetBasicAuth(os.Getenv(EnvKeySumoAccessID), os.Getenv(EnvKeySumoAccessKey)).Get(endpoint)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("Error: %s", resp.Status())
	}

	respBody := struct {
		Data  []MonitorConnections `json:"data,omitempty"`
		Token string               `json:"token,omitempty"`
	}{
		Data: []MonitorConnections{},
	}

	json.Unmarshal(resp.Body(), &respBody)

	return respBody.Data, nil
}
