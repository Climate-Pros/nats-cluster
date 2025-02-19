package main

import (
	"context"
	"fmt"
	"github.com/lucsky/cuid"
	"os"
	"os/exec"
	"sort"
	"syscall"
	"text/template"
	"time"

	_ "embed"
	"github.com/fly-apps/nats-cluster/pkg/privnet"
	"github.com/fly-apps/nats-cluster/pkg/supervisor"
)

func main() {
	natsVars, err := initNatsConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	svisor := supervisor.New("flynats", 5*time.Minute)

	svisor.AddProcess(
		"exporter",
		"nats-exporter -varz 'http://fly-local-6pn:8222'",
		supervisor.WithRestart(0, 1*time.Second),
	)

	svisor.AddProcess(
		"nats-server",
		"nats-server -c /etc/nats.conf --logtime=false",
		supervisor.WithRestart(0, 1*time.Second),
	)

	go watchNatsConfig(natsVars)

	svisor.StopOnSignal(syscall.SIGINT, syscall.SIGTERM)

	svisor.StartHttpListener()

	err = svisor.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type FlyEnv struct {
	Host            string
	AppName         string
	Region          string
	GatewayRegions  []string
	ServerName      string
	Timestamp       time.Time
	AppAccountName  string
	OperatorName    string
	OperatorToken   string
	SystemAccount   string
	ResolverPayload string
	StoreDir        string
	MaxMemory       string
	MaxDisk         string
}

//go:embed nats.conf.tmpl
var tmplRaw string

func watchNatsConfig(vars FlyEnv) {
	fmt.Println("Starting ticker")
	ticker := time.NewTicker(5 * time.Second)
	var lastReload time.Time

	go func() {
		for {
			for range ticker.C {
				newVars, err := natsConfigVars()

				if err != nil {
					fmt.Printf("error getting nats config vars: %v", err)
					continue
				}
				if stringSlicesEqual(vars.GatewayRegions, newVars.GatewayRegions) {
					// noop, nothing changed
					//fmt.Println("No change in regions")
					continue
				}

				cooloff := lastReload.Add(15 * time.Second)
				if time.Now().Before(cooloff) {
					fmt.Println("Regions changed, but cooloff period not expired")
					continue
				}

				err = writeNatsConfig(newVars)
				if err != nil {
					fmt.Printf("error writing nats config: %v", err)
				}

				cmd := exec.Command(
					"nats-server",
					"--signal",
					"stop=/var/run/nats-server.pid",
				)
				fmt.Printf("Reloading nats: \n\t%v\n\t%v\n", vars.GatewayRegions, newVars.GatewayRegions)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err = cmd.Run()
				if err != nil {
					fmt.Printf("Command finished with error: %v", err)
				}
				vars = newVars
				lastReload = time.Now()
			}
		}
	}()

	fmt.Println("ticker fn return")
}

func natsConfigVars() (FlyEnv, error) {
	var err error

	host := "fly-local-6pn"
	appName := os.Getenv("FLY_APP_NAME")
	storeDir := os.Getenv("NATS_STORE_DIR")

	var regions []string

	if appName != "" {
		regions, err = privnet.GetRegions(context.Background(), appName)
	} else {
		// defaults for local exec
		host = "localhost"
		appName = "local"
		regions = []string{"local"}
	}

	// easier to compare
	sort.Strings(regions)

	region := os.Getenv("FLY_REGION")
	if region == "" {
		region = "local"
	}

	serverNamePath := storeDir + "/server_name"
	var serverName = []byte(region)
	serverName, err = os.ReadFile(serverNamePath)
	if err != nil {
		fmt.Printf("Servername not stored. Creating file:  %s", serverNamePath)
		_, err = os.Create(serverNamePath)
		if err != nil {
			fmt.Printf("Error creating file %s", serverNamePath)
			return FlyEnv{}, err
		}
	}
	err = cuid.IsCuid(string(serverName))
	if err != nil {
		serverName = []byte(region + "_" + cuid.New())
		err = os.WriteFile(serverNamePath, []byte(serverName), 0666)
		if err != nil {
			fmt.Println("Failed to save", string(serverName), "to file", serverNamePath)
			fmt.Println(err)
			return FlyEnv{}, err
		}
	}

	vars := FlyEnv{
		AppName:         appName,
		Region:          region,
		GatewayRegions:  regions,
		Host:            host,
		ServerName:      string(serverName),
		Timestamp:       time.Now(),
		StoreDir:        storeDir,
		MaxMemory:       os.Getenv("NATS_MAX_MEMORY"),
		MaxDisk:         os.Getenv("NATS_MAX_DISK"),
		AppAccountName:  os.Getenv("NATS_APP_USER_NAME"),
		OperatorName:    os.Getenv("NATS_OPERATOR_NAME"),
		OperatorToken:   os.Getenv("NATS_OPERATOR_TOKEN"),
		SystemAccount:   os.Getenv("NATS_SYSTEM_ACCOUNT"),
		ResolverPayload: os.Getenv("NATS_RESOLVER_PAYLOAD"),
	}
	if err != nil {
		return FlyEnv{}, err
	}
	return vars, nil
}

func initNatsConfig() (FlyEnv, error) {
	vars, err := natsConfigVars()
	if err != nil {
		return vars, err
	}
	err = writeNatsConfig(vars)

	if err != nil {
		return vars, err
	}

	return vars, nil
}

func writeNatsConfig(vars FlyEnv) error {
	tmpl, err := template.New("conf").Parse(tmplRaw)

	if err != nil {
		return err
	}

	f, err := os.Create("/etc/nats.conf")

	if err != nil {
		return err
	}

	err = tmpl.Execute(f, vars)

	if err != nil {
		return err
	}

	return nil
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
