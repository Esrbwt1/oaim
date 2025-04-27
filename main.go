package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: oaim <command> [flags]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	// shift args so flag package sees only command-specific flags
	os.Args = append(os.Args[:1], os.Args[2:]...)
	switch cmd {
	case "init":
		cmdInit()
	case "discover":
		cmdDiscover()
	case "run-agent":
		cmdRunAgent()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}

}

func cmdInit() {
	spec := `version: "0.1"
domains:
  - name: discovery
    description: "How agents find each other"
  - name: negotiation
    description: "Task proposals & acceptances"
  - name: execution
    description: "Data exchange & status"
  - name: billing
    description: "Micropayments & revenue reporting"`
	path := filepath.Join(".", "spec.yaml")
	if _, err := os.Stat(path); err == nil {
		fmt.Println("spec.yaml already exists; aborting.")
		os.Exit(1)
	}
	err := os.WriteFile(path, []byte(spec), 0644)
	if err != nil {
		fmt.Println("Error writing spec.yaml:", err)
		os.Exit(1)
	}
	fmt.Println("Initialized spec.yaml")
}

func cmdRunAgent() {
	// Flags
	port := flag.Int("port", 8080, "Port to serve on")
	pingURL := flag.String("ping", "", "URL of peer to ping (overrides auto-discovery)")
	flag.Parse()

	// If explicit ping flag is given, act as client only
	if *pingURL != "" {
		resp, err := http.Get(*pingURL)
		if err != nil {
			log.Fatalf("Ping failed: %v", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Received response from peer: %s\n", string(body))
		return
	}

	// Otherwise, start as server
	// First, auto-discover bootstrap peers and ping them
	peers, err := loadBootstrapPeers()
	if err != nil {
		log.Printf("Warning: could not load bootstrap peers: %v\n", err)
	} else {
		fmt.Println("🔍 Auto-discovering peers...")
		for _, peer := range peers {
			fmt.Printf("Pinging %s … ", peer)
			resp, err := http.Get(peer)
			if err != nil {
				fmt.Println("failed:", err)
			} else {
				body, _ := ioutil.ReadAll(resp.Body)
				fmt.Println(string(body))
			}
		}
	}

	// Now start HTTP server for other agents to ping
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})
	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🟢 OAIM Agent listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func cmdDiscover() {
	// Read spec.yaml
	data, err := os.ReadFile("spec.yaml")
	if err != nil {
		log.Fatalf("Error reading spec.yaml: %v", err)
	}

	// Parse YAML (using minimal external dependency)
	type Spec struct {
		Bootstrap struct {
			Peers []string `yaml:"peers"`
		} `yaml:"bootstrap"`
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		log.Fatalf("Error parsing spec.yaml: %v", err)
	}

	// Print peer list
	fmt.Println("Discovered bootstrap peers:")
	for _, peer := range spec.Bootstrap.Peers {
		fmt.Printf("  • %s\n", peer)
	}
}

// loadBootstrapPeers reads spec.yaml and returns the list of bootstrap peer URLs.
func loadBootstrapPeers() ([]string, error) {
	data, err := os.ReadFile("spec.yaml")
	if err != nil {
		return nil, err
	}
	type Spec struct {
		Bootstrap struct {
			Peers []string `yaml:"peers"`
		} `yaml:"bootstrap"`
	}
	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	return spec.Bootstrap.Peers, nil
}
