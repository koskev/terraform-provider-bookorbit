package bookrobitnet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func (c *BookorbitClient) Token() (string, error) {
	token := os.Getenv("BOOKORBIT_TOKEN")

	if len(token) != 0 {
		return token, nil
	}

	if len(c.username) == 0 || len(c.password) == 0 {
		return "", fmt.Errorf("the BOOKORBIT_TOKEN environment variable is empty and no username/password was provided in the config")
	}

	body := fmt.Sprintf("{\"username\":\"%s\",\"password\":\"%s\"}", c.username, c.password)
	req, err := http.NewRequest("POST", c.getUrl("api/v1/auth/login"), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var loginResponse struct {
		Token string `json:"accessToken"`
	}

	resp, err := c.SendHTTPRequest(req)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(resp), &loginResponse)
	if err != nil {
		return "", err
	}

	return loginResponse.Token, nil
}

type BookorbitClient struct {
	baseUrl  string
	username string
	password string
	skipTLS  bool
}

func New(url string, username string, password string, skipTLS bool) BookorbitClient {
	return BookorbitClient{
		baseUrl:  url,
		username: username,
		password: password,
		skipTLS:  skipTLS,
	}
}

func (c *BookorbitClient) getUrl(endpoint string) string {
	return fmt.Sprintf("%s/%s", c.baseUrl, endpoint)
}

func (c *BookorbitClient) SendRequest(method string, endpoint string, body []byte, additional_headers map[string]string) (string, error) {
	url := c.getUrl(endpoint)
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	return c.SendHTTPRequest(req)
}

func (c *BookorbitClient) SendAuthenticatedRequest(method string, endpoint string, body []byte) (string, error) {
	token, err := c.Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}
	return c.SendRequest(method, endpoint, body, map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	})
}

func (c *BookorbitClient) SendHTTPRequest(req *http.Request) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.skipTLS},
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Warn("Unable to close body", "err", err)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return string(respBody), fmt.Errorf("got invalid status code: %d", resp.StatusCode)
	}

	return string(respBody), nil
}

func (c *BookorbitClient) Get(endpoint string) (string, error) {
	return c.SendAuthenticatedRequest("GET", endpoint, nil)
}
