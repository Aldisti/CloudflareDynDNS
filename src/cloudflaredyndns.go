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
	ctx := buildCtx()

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

func routine(ctx *Context) bool {

	if ctx.Env.Cooldown >= 0 && time.Since(ctx.LastFailure) > ctx.Env.Cooldown {
		ctx.resetFailures()
		fmt.Println("Failures reset") // debug
	} else if ctx.Env.MaxFails >= 0 && ctx.Failures > ctx.Env.MaxFails {
		fmt.Println("Reached maximum number of failures, aborting") // info
		return false
	}

	ip, err := getCurrentIp()
	if err != nil {
		ctx.addFailure()
		fmt.Println(err) // debug
		return true
	}

	for i, record := range ctx.Records {
		record, err = cloudflare.UpdateRecord(record.Name, record.ID, ip)
		if err != nil {
			ctx.addFailure()
			fmt.Println(err) // debug
		} else {
			ctx.CurrentIP = ip
			ctx.Records[i] = record
			fmt.Printf("Record %s updated with new ip: %s\n", record.Name, ip) // info
		}
	}

	return true
}


func buildCtx() (Context) {
	ctx := Context{
		Failures: 0,
		Records: make([]cloudflare.Record, 0),
	}

	if env, err := internal.GetEnvSafe(); err != nil {
		panic(err)
	} else {
		ctx.Env = env
	}

	if ip, err := getCurrentIp(); err != nil {
		panic(err)
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
			panic(fmt.Errorf("BuildCtx: %s", err))
		}
		ctx.Records = append(ctx.Records, record)
		fmt.Printf("Record %s created\n", domain)
	}

	fmt.Println("Context successfully built")
	return ctx
}

func (ctx *Context) addFailure() {
	ctx.Failures++
	ctx.LastFailure = time.Now()
}

func (ctx *Context) resetFailures() {
	ctx.Failures = 0
	ctx.LastFailure = time.Now()
}

func getCurrentIp() (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", fmt.Errorf("NewRequest failed: %s", err)
	}

	env := internal.GetEnv()
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
