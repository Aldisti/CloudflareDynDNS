package poller

import (
	"fmt"
	"strings"
	"time"

	"github.com/Aldisti/CloudflareDynDNS/cloudflare"
	"github.com/Aldisti/CloudflareDynDNS/common"
	"github.com/Aldisti/CloudflareDynDNS/config"
)

type Context struct {
	// Settings

	Domains     []string
	Interval    int
	MaxFailures int
	Cooldown    time.Duration

	// Variables

	Records     []cloudflare.Record
	CurrentIP   string
	LastUpdate  time.Time
	Failures    int
	LastFailure time.Time
}

func (ctx *Context) addFailure() {
	ctx.Failures++
	ctx.LastFailure = time.Now()
	fmt.Println("Failures reset") // debug
}

func (ctx *Context) resetFailures() {
	ctx.Failures = 0
	ctx.LastFailure = time.Now()
}

func Run(env *config.Environment) {
	ctx := buildCtx(env)

	done := make(chan bool)
	ticker := time.NewTicker(time.Duration(ctx.Interval) * time.Second)
	defer ticker.Stop()

	fmt.Println("Starting POLLER mode") // info

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if !routine(&ctx) {
				return
			}
		}
	}
}

func routine(ctx *Context) bool {

	if ctx.Cooldown > 0 && time.Since(ctx.LastFailure) > ctx.Cooldown {
		ctx.resetFailures()
	} else if ctx.MaxFailures >= 0 && ctx.Failures > ctx.MaxFailures {
		fmt.Println("Reached maximum number of failures, aborting") // warning
		return false
	}

	ip, err := cloudflare.GetCurrentIp()
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

func buildCtx(env *config.Environment) Context {
	ctx := Context{
		Failures: 0,
		Records:  make([]cloudflare.Record, 0),
	}

	ctx.Domains = common.Filter(strings.Split(env.Domains, ","), common.IsNotBlank)
	ctx.Interval = common.GetIntUnsafe(env.Interval, "interval")
	ctx.MaxFailures = common.GetIntUnsafe(env.MaxFails, "max failures")
	ctx.Cooldown = time.Duration(common.GetIntUnsafe(env.Cooldown, "cooldown")) * time.Second

	if err := validateEnv(ctx); err != nil {
		panic(fmt.Errorf("Poller::buildCtx: %v", err))
	}

	if _, err := cloudflare.GetCurrentIp(); err != nil {
		panic(fmt.Errorf("Cannot retrieve current public ip: %v", err))
	}

	for _, domain := range ctx.Domains {
		if record, err := getOrCreateRecord(&ctx, domain); err != nil {
			panic(fmt.Errorf("Poller::buildCtx: %v", err))
		} else {
			ctx.Records = append(ctx.Records, record)
		}
	}

	fmt.Println("Poller context successfully built") // info
	return ctx
}

func getOrCreateRecord(ctx *Context, domain string) (cloudflare.Record, error) {
	record, ok, err := cloudflare.GetFirstRecord(domain, "A")
	if ok {
		return record, nil
	}
	if err != nil {
		fmt.Printf("Error while searching for %s: %s\n", domain, err) // warning
	}
	record = cloudflare.Record{
		Name:    domain,
		Type:    "A",
		Content: ctx.CurrentIP,
		TTL:     max(ctx.Interval/2, 60), // TODO: move these values into env vars
		Proxied: false,
		Comment: "Created by github.com/aldisti/CloudflareDynDNS",
	}
	record, err = cloudflare.CreateRecord(record)
	if err != nil {
		return record, fmt.Errorf("getOrCreateRecord: %v", err) // error
	}
	fmt.Println("Created record:", domain) // info
	return record, nil
}

func validateEnv(ctx Context) error {
	if len(ctx.Domains) == 0 {
		return fmt.Errorf("Configure at least 1 domain to update")
	}
	if ctx.Interval <= 0 {
		return fmt.Errorf("Interval must be greater than 0")
	}
	return nil
}
