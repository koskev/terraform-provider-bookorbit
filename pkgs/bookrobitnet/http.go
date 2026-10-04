package bookrobitnet

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func Token() (string, error) {
	token := os.Getenv("BOOKORBIT_TOKEN")

	if len(token) == 0 {
		return "", fmt.Errorf("the BOOKORBIT_TOKEN environment variable is empty")
	}

	return token, nil
}

type BookorbitClient struct {
	baseUrl string
}

func New(url string) BookorbitClient {
	return BookorbitClient{
		baseUrl: url,
	}
}

func (c *BookorbitClient) getUrl(endpoint string) string {
	return fmt.Sprintf("%s/%s", c.baseUrl, endpoint)
}

func (c *BookorbitClient) SendRequest(method string, endpoint string, body []byte) (string, error) {
	url := c.getUrl(endpoint)
	bodyReader := bytes.NewReader(body)
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	token, err := Token()
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
	return c.SendRequest("GET", endpoint, nil)
}
