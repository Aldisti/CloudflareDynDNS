package poller

import (
	"fmt"
	"time"

	"github.com/Aldisti/CloudflareDynDNS/cloudflare"
	"github.com/Aldisti/CloudflareDynDNS/common"
	"github.com/Aldisti/CloudflareDynDNS/config"
)

type Context struct {
	Env         *config.Environment
	Records      []cloudflare.Record
	CurrentIP   string
	LastUpdate  time.Time
	Failures    int
	LastFailure time.Time
}

func Run() {
	ctx := buildCtx()

	done := make(chan bool)
	ticker := time.NewTicker(time.Duration(ctx.Env.Interval) * time.Second)
	defer ticker.Stop()

	fmt.Println("Starting POLLER mode") // info

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

	ip, err := common.GetCurrentIp()
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

	if env, err := config.GetEnvSafe(); err != nil {
		panic(err)
	} else {
		ctx.Env = env
	}

	if ip, err := common.GetCurrentIp(); err != nil {
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
