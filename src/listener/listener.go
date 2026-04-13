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

var credentials = make(map[string]string)

type Context struct {
	Port    int
	Address string
}

func Run(env *config.Environment) {
	ctx := buildCtx(env)

	http.HandleFunc("/update", handleUpdate)

	fmt.Printf("Starting server at %s:%d\n", ctx.Address, ctx.Port)

	addr := fmt.Sprintf("%s:%d", ctx.Address, ctx.Port)
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
	ip, err := cloudflare.GetCurrentIp()
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
			fmt.Printf("Record %s updated with new ip %s\n", domain, ip) // info
		}
	}
	return nil
}

func buildCtx(env *config.Environment) Context {
	ctx := Context{}

	if common.IsBlank(env.Username) {
		panic(fmt.Errorf("Missing env var: %s", config.ENV_USERNAME))
	}
	if common.IsBlank(env.Password) {
		panic(fmt.Errorf("Missing env var: %s", config.ENV_PASSWORD))
	}
	credentials[env.Username] = env.Password

	ctx.Port = common.GetIntUnsafe(env.Port, "port")
	ctx.Address = env.Address

	if err := validateEnv(ctx); err != nil {
		panic(fmt.Errorf("Listener::buildCtx: %v", err))
	}

	return ctx
}

func validateEnv(ctx Context) error {
	if ctx.Port <= 0 {
		return fmt.Errorf("Port must be greater than 0")
	} else if ctx.Port >= 65536 {
		return fmt.Errorf("Port must be lower than 65536")
	}

	if common.IsBlank(ctx.Address) {
		return fmt.Errorf("Address cannot be blank")
	}
	return nil
}

func compareCredentials(username, password string) bool {
	if p, ok := credentials[username]; !ok {
		return false
	} else {
		return p == password
	}
}
