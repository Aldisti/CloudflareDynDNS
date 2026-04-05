package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	ENV_ZONE_ID   = "ZONE_ID"
	ENV_RECORD_ID = "RECORD_ID"
	ENV_API_TOKEN = "API_TOKEN"
	ENV_API_URL   = "API_URL"
)

type Environment struct {
	ZoneId   string
	RecordId string
	ApiToken string
	ApiUrl   string
}

type CloudflareMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CloudflareResponse struct {
	Success  bool                `json:"success"`
	Errors   []CloudflareMessage `json:"errors"`
	Messages []CloudflareMessage `json:"messages"`
}

func main() {
	env, err := loadEnvironment()
	if err != nil {
		fmt.Println(err)
		return
	}
	ip, err := getCurrentIp()
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = updateRecord(env, ip); err != nil {
		fmt.Println(err)
	}
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func loadEnvironment() (Environment, error) {
	env := Environment{
		ApiUrl: "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s",
	}
	if s, ok := os.LookupEnv(ENV_ZONE_ID); !ok || isBlank(s) {
		return env, fmt.Errorf("Error: missing env variable: %s", ENV_ZONE_ID)
	} else {
		env.ZoneId = s
	}
	if s, ok := os.LookupEnv(ENV_RECORD_ID); !ok || isBlank(s) {
		return env, fmt.Errorf("Error: missing env variable: %s", ENV_RECORD_ID)
	} else {
		env.RecordId = s
	}
	if s, ok := os.LookupEnv(ENV_API_TOKEN); !ok || isBlank(s) {
		return env, fmt.Errorf("Error: missing env variable: %s", ENV_API_TOKEN)
	} else {
		env.ApiToken = s
	}
	if s, ok := os.LookupEnv(ENV_API_URL); !ok || isBlank(s) {
		fmt.Printf("Warning: using %s default value\n", ENV_API_URL)
	} else {
		env.ApiUrl = s
	}
	return env, nil
}

func getCurrentIp() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", fmt.Errorf("NewRequest failed: %s", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("DefaultClient.Do failed: %s", err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("Couldn't read response: %s", err)
	}
	if res.StatusCode == 200 {
		return string(resBody), nil
	} else {
		return "", fmt.Errorf("Received status code %d: %s", res.StatusCode, resBody)
	}
}

func updateRecord(env Environment, newIp string) error {

	url := fmt.Sprintf(env.ApiUrl, env.ZoneId, env.RecordId)
	body := fmt.Sprintf(`{"content": "%s"}`, newIp)

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader([]byte(body)))
	if err != nil { // create new request
		return fmt.Errorf("Couldn't create PATCH request: %s", err)
	}

	// add headers to request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.ApiToken))

	client := http.Client{ // create custom http client
		Timeout: 1 * time.Second,
	}

	res, err := client.Do(req) // make request
	if err != nil {
		return fmt.Errorf("PATCH request failed: %s", err)
	}

	resBody, err := io.ReadAll(res.Body) // read response body
	if err != nil {
		return fmt.Errorf("Couldn't read response: %s", err)
	}

	if res.StatusCode != http.StatusOK { // check response status
		return fmt.Errorf("Received status code %d: %s", res.StatusCode, string(resBody))
	}

	var response CloudflareResponse // unmarshall response body into struct
	if err = json.Unmarshal(resBody, &response); err != nil {
		return fmt.Errorf("Couldn't parse cloudflare response: %s", err)
	}

	if response.Success { // check cloudflare response status
		fmt.Println("Record successfully updated with new IP:", newIp)
		return nil
	}

	return fmt.Errorf("Error while updating record: %s", string(resBody))
}

func Ticker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Ticking", t)
			}
		}
	}()

	time.Sleep(30 * time.Second) // Run for 30 seconds
	done <- true
}
