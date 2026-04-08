package listener

import (
	"fmt"
	"net/http"

	"github.com/Aldisti/CloudflareDynDNS/cloudflare"
	"github.com/Aldisti/CloudflareDynDNS/common"
	"github.com/Aldisti/CloudflareDynDNS/config"
)

const (
	PARAM_HOSTNAME = "hostname"
)

type Context struct {
	Env      *config.Environment
}

func Run() {
	ctx := buildCtx()

	http.HandleFunc("/update", handleUpdate)

	fmt.Printf("Starting server on port %d\n", ctx.Env.Port)

	addr := fmt.Sprintf("0.0.0.0:%d", ctx.Env.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Error: failed to ListenAndServe:", err)
	}
}

func handleUpdate(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if ok && compareCredentials(username, password) {
		q := r.URL.Query()
		if !q.Has(PARAM_HOSTNAME) {
			fmt.Printf("Missing %s parameter in request\n", PARAM_HOSTNAME)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		hostnames := q[PARAM_HOSTNAME]
		if err := updateHostnames(hostnames); err != nil {
			fmt.Printf("Error while updating hostnames %v: %s\n", hostnames, err)
			http.Error(w, "Error while updating hostnames", http.StatusInternalServerError)
		}
		return
	}
	fmt.Printf("Received invalid credentials for username: '%s'\n", username)
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func updateHostnames(domains []string) error {
	ip, err := common.GetCurrentIp()
	if err != nil {
		return err
	}
	for _, domain := range domains {
		record, ok, err := cloudflare.GetFirstRecord(domain, "A")
		if err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("Record %s not found", domain)
		}
		if _, err := cloudflare.UpdateRecord(domain, record.ID, ip); err != nil {
			return err
		} else {
			fmt.Printf("Record %s updated with new ip %s\n", domain, ip)
		}
	}
	return nil
}

func buildCtx() Context {
	ctx := Context{
		Env: config.GetEnv(),
	}

	if common.IsBlank(ctx.Env.Username) {
		panic(fmt.Errorf("Error: missing env variable %s", config.ENV_USERNAME))
	}
	if common.IsBlank(ctx.Env.Password) {
		panic(fmt.Errorf("Error: missing env variable %s", config.ENV_PASSWORD))
	}

	return ctx
}

func compareCredentials(username, password string) bool {
	env := config.GetEnv()
	return env.Username == username && env.Password == password
}
