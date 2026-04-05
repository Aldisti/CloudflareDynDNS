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
	Record      cloudflare.Record
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

	ticker := time.NewTicker(time.Duration(ctx.Env.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
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

	if record, ok, err := cloudflare.GetFirstRecord(ctx.Env.Domain, "A"); err != nil || !ok {
		if err != nil {
			fmt.Println(err)
		}
		record = cloudflare.Record{
			Name:    ctx.Env.Domain,
			Type:    "A",
			Content: ctx.CurrentIP,
			TTL:     max(ctx.Env.Interval*2, 60),
			Proxied: false,
			Comment: "Created by CloudflareDynDNS",
		}
		record, err = cloudflare.CreateRecord(record)
		if err != nil {
			return ctx, fmt.Errorf("BuildCtx: %s", err)
		}
		ctx.Record = record
		fmt.Println("Record created successfully")
	} else {
		ctx.Record = record
	}

	return ctx, nil
}

func routine(ctx *Context) (bool) {
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

	if ip == ctx.CurrentIP && ip == ctx.Record.Content {
		fmt.Println("Record already up to date, skipping")
		return true
	}

	if err := cloudflare.UpdateRecord(ctx.Record.ID, ip); err != nil {
		fmt.Println(err)
		addFailure(ctx)
	} else {
		fmt.Println("Record updated with new ip:", ip)
		ctx.CurrentIP = ip
		ctx.Record.Content = ip
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

	client := http.Client{Timeout: 500 * time.Millisecond}
	res, err := client.Do(req)
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
