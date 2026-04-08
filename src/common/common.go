package common

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Aldisti/CloudflareDynDNS/config"
)

func GetCurrentIp() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", fmt.Errorf("NewRequest failed: %s", err) // warn
	}

	env := config.GetEnv()
	client := http.Client{Timeout: env.Timeout}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Couldn't make request: %s", err)
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

func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}
