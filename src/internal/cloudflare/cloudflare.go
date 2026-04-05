package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Aldisti/CloudflareDynDNS/internal"
)

const (
	LIST_RECORDS   = "https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=%s"
	CREATE_RECORDS = "https://api.cloudflare.com/client/v4/zones/%s/dns_records"
	UPDATE_RECORD  = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
)

type Record struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Comment string `json:"comment"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MultiResponse struct {
	Success  bool      `json:"success"`
	Errors   []Message `json:"errors"`
	Messages []Message `json:"messages"`
	Results  []Record  `json:"result"`
}

type SingleResponse struct {
	Success  bool      `json:"success"`
	Errors   []Message `json:"errors"`
	Messages []Message `json:"messages"`
	Result   Record    `json:"result"`
}

/*
 * Main API methods
 */

func GetFirstRecord(name, tipe string) (Record, bool, error) {
	var record Record
	env := internal.GetEnv()

	url := fmt.Sprintf(LIST_RECORDS, env.ZoneId, name, tipe)
	req, err := buildRequest(http.MethodGet, url, "")
	if err != nil {
		return record, false, fmt.Errorf("GetFirstRecord: %s", err)
	}

	var res MultiResponse
	if err := makeRequest(req, &res); err != nil {
		return record, false, fmt.Errorf("GetFirstRecord: %s", err)
	} else if !res.Success {
		return record, false, fmt.Errorf("GetFirstRecord: Couldn't search record '%s' of type '%s'", name, tipe)
	}

	if len(res.Results) > 0 {
		return res.Results[0], true, nil
	} else {
		return record, false, nil
	}
}

func CreateRecord(record Record) (Record, error) {
	env := internal.GetEnv()

	url := fmt.Sprintf(CREATE_RECORDS, env.ZoneId)
	body, err := json.Marshal(record)
	if err != nil {
		return record, fmt.Errorf("Failed to marshal record: %s", err)
	}

	req, err := buildRequest(http.MethodPost, url, string(body))
	if err != nil {
		return record, fmt.Errorf("CreateRecord: %s", err)
	}

	var res SingleResponse
	if err = makeRequest(req, &res); err != nil {
		return record, fmt.Errorf("CreateRecord: %s", err)
	} else if !res.Success {
		return record, fmt.Errorf("CreateRecord: couldn't create new record")
	} else {
		return res.Result, nil
	}
}

func UpdateRecord(recordId, content string) error {
	env := internal.GetEnv()

	url := fmt.Sprintf(UPDATE_RECORD, env.ZoneId, recordId)
	body := fmt.Sprintf(`{"content":"%s"}`, content)

	req, err := buildRequest(http.MethodPatch, url, body)
	if err != nil {
		return fmt.Errorf("UpdateRecord: %s", err)
	}

	var res SingleResponse
	if err = makeRequest(req, &res); err != nil {
		return fmt.Errorf("UpdateRecord: %s", err)
	}

	if res.Success {
		return nil
	} else {
		return fmt.Errorf("UpdateRecord: Couldn't update record")
	}

}

/*
 * Internal functions
 */

func buildRequest(method, url, body string) (*http.Request, error) {
	env := internal.GetEnv()

	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewReader([]byte(body))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to build request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.ApiToken))

	return req, nil
}

func makeRequest(req *http.Request, response any) error {
	env := internal.GetEnv()
	client := http.Client{
		Timeout: time.Duration(env.Timeout) * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %s", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("Failed read response: %s", err)
	}

	err = json.Unmarshal(body, response)
	if err != nil {
		fmt.Println(string(body))
		return fmt.Errorf("Failed to unmarshall response: %s", err)
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("Received status code %d: %s", res.StatusCode, string(body))
	}

	return nil
}
