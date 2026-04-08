package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Aldisti/CloudflareDynDNS/config"
)

const (
	LIST_ZONES     = "https://api.cloudflare.com/client/v4/zones"
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

type Zone struct {
	ID                  string   `json:"id"`
	DevelopmentMode     int      `json:"development_mode"`
	Name                string   `json:"name"`
	OriginalDnshost     string   `json:"original_dnshost"`
	OriginalRegistrar   string   `json:"original_registrar"`
	CnameSuffix         string   `json:"cname_suffix"`
	Status              string   `json:"status"`
	Type                string   `json:"type"`
	Paused              bool     `json:"paused"`
	NameServers         []string `json:"name_servers"`
	OriginalNameServers []string `json:"original_name_servers"`
	Permissions         []string `json:"permissions"`
	VanityNameServers   []string `json:"vanity_name_servers"`
}

type Message struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response[T any] struct {
	Success  bool      `json:"success"`
	Errors   []Message `json:"errors"`
	Messages []Message `json:"messages"`
	Result   T         `json:"result"`
}

var (
	zones = make(map[string]Zone)
)

/*
 * Main API methods
 */

func GetFirstRecord(name, tipe string) (Record, bool, error) {
	var record Record

	zone, err := FindMatchingZone(name)
	if err != nil {
		return record, false, err
	}

	url := fmt.Sprintf(LIST_RECORDS, zone.ID, name, tipe)
	req, err := buildRequest(http.MethodGet, url, "")
	if err != nil {
		return record, false, fmt.Errorf("GetFirstRecord: %s", err)
	}

	var res Response[[]Record]
	if err := makeRequest(req, &res); err != nil {
		return record, false, fmt.Errorf("GetFirstRecord: %s", err)
	} else if !res.Success {
		return record, false, fmt.Errorf("GetFirstRecord: Couldn't search record '%s' of type '%s'", name, tipe)
	}

	if len(res.Result) > 0 {
		return res.Result[0], true, nil
	} else {
		return record, false, nil
	}
}

func CreateRecord(record Record) (Record, error) {
	zone, err := FindMatchingZone(record.Name)
	if err != nil {
		return record, err
	}

	url := fmt.Sprintf(CREATE_RECORDS, zone.ID)
	body, err := json.Marshal(record)
	if err != nil {
		return record, fmt.Errorf("Failed to marshal record: %s", err)
	}

	req, err := buildRequest(http.MethodPost, url, string(body))
	if err != nil {
		return record, fmt.Errorf("CreateRecord: %s", err)
	}

	var res Response[Record]
	if err = makeRequest(req, &res); err != nil {
		return record, fmt.Errorf("CreateRecord: %s", err)
	} else if !res.Success {
		return record, fmt.Errorf("CreateRecord: couldn't create new record")
	} else {
		return res.Result, nil
	}
}

func UpdateRecord(domain, recordId, content string) (Record, error) {
	var record Record
	zone, err := FindMatchingZone(domain)
	if err != nil {
		return record, err
	}

	url := fmt.Sprintf(UPDATE_RECORD, zone.ID, recordId)
	body := fmt.Sprintf(`{"content":"%s"}`, content)
	req, err := buildRequest(http.MethodPatch, url, body)
	if err != nil {
		return record, fmt.Errorf("UpdateRecord: %s", err)
	}

	var res Response[Record]
	if err = makeRequest(req, &res); err != nil {
		return record, fmt.Errorf("UpdateRecord: %s", err)
	}

	if res.Success {
		return res.Result, nil
	} else {
		return record, fmt.Errorf("UpdateRecord: Couldn't update record")
	}
}

func FindMatchingZone(domain string) (Zone, error) {

	if len(zones) == 0 {
		if err := SetupZones(); err != nil {
			return Zone{}, fmt.Errorf("FindMatchingZone: %s", err)
		}
	}

	for name, zone := range zones {
		if strings.HasSuffix(domain, name) {
			return zone, nil
		}
	}

	return Zone{}, fmt.Errorf("FindZone: didn't find any zone matching %s", domain)
}

// SetupZones is a setup function that loads in memory all the available zones.
//
// Should be used only once at the startup of the program.
func SetupZones() error {
	if len(zones) > 0 {
		return nil
	}

	req, err := buildRequest(http.MethodGet, LIST_ZONES, "")
	if err != nil {
		return fmt.Errorf("SetupZones: %s", err)
	}

	var res Response[[]Zone]
	if err = makeRequest(req, &res); err != nil {
		return fmt.Errorf("SetupZones: %s", err)
	}

	if !res.Success {
		return fmt.Errorf("SetupZones: failed")
	} else if len(res.Result) == 0 {
		return fmt.Errorf("SetupZones: no zone found")
	}

	for _, zone := range res.Result {
		zones[zone.Name] = zone
	}

	return nil
}

/*
 * Internal functions
 */

func buildRequest(method, url, body string) (*http.Request, error) {
	env := config.GetEnv()

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
	env := config.GetEnv()
	client := http.Client{
		Timeout: env.Timeout,
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
