package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Aldisti/CloudflareDynDNS/internal"
	"github.com/Aldisti/CloudflareDynDNS/internal/cloudflare"
)

type Context struct {
	Env         *internal.Environment
	Records      []cloudflare.Record
	CurrentIP   string
	LastUpdate  time.Time
	Failures    int
	LastFailure time.Time
}

func main() {
	ctx, err := buildCtx()
	if err != nil {
		fmt.Println(err)
		return
	}

	done := make(chan bool)
	ticker := time.NewTicker(time.Duration(ctx.Env.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <- done:
			return
		case <-ticker.C:
			if !routine(&ctx) {
				return
			}
		}
	}
}

func buildCtx() (Context, error) {
	ctx := Context{
		Failures: 0,
		Records: make([]cloudflare.Record, 0),
	}

	if env, err := internal.GetEnvSafe(); err != nil {
		return ctx, err
	} else {
		ctx.Env = env
	}

	if ip, err := getCurrentIp(); err != nil {
		return ctx, err
	} else {
		ctx.CurrentIP = ip
	}

	for _, domain := range ctx.Env.Domains {
		record, ok, err := cloudflare.GetFirstRecord(domain, "A")
		if ok {
			ctx.Records = append(ctx.Records, record)
			continue
		}
		if err != nil {
			fmt.Printf("Error while searching for %s: %s\n", domain, err)
		}
		record = cloudflare.Record{
			Name: domain,
			Type: "A",
			Content: ctx.CurrentIP,
			TTL: max(ctx.Env.Interval * 2, 60),
			Proxied: false,
			Comment: "Created by github.com/aldisti/CloudflareDynDNS",
		}
		record, err = cloudflare.CreateRecord(record)
		if err != nil {
			return ctx, fmt.Errorf("BuildCtx: %s", err)
		}
		ctx.Records = append(ctx.Records, record)
		fmt.Printf("Record %s created\n", domain)
	}

	return ctx, nil
}

func routine(ctx *Context) bool {
	if ctx.Failures > ctx.Env.MaxFails {
		fmt.Println("Reached maximum number of failures, aborting")
		return false
	}

	ip, err := getCurrentIp()
	if err != nil {
		fmt.Println(err)
		addFailure(ctx)
		return true
	}

	if ip == ctx.CurrentIP {
		fmt.Println("IP didn't change, skipping")
		return true
	}

	for i, record := range ctx.Records {
		record, err = cloudflare.UpdateRecord(record.Name, record.ID, ip)
		if err != nil {
			addFailure(ctx)
			fmt.Println(err)
		} else {
			ctx.CurrentIP = ip
			ctx.Records[i] = record
			fmt.Printf("Record %s updated with new ip: %s\n", record.Name, ip)
		}
	}

	return true
}

func addFailure(ctx *Context) {
	ctx.Failures++
	ctx.LastFailure = time.Now()
}

func getCurrentIp() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", fmt.Errorf("NewRequest failed: %s", err)
	}

	env := internal.GetEnv()
	client := http.Client{Timeout: time.Duration(env.Timeout) * time.Second}
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
