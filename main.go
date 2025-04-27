package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	// Define flags
	port := flag.Int("port", 8080, "Port to serve on")
	peer := flag.String("ping", "", "URL of peer to ping (e.g. http://localhost:8081/ping)")
	flag.Parse()

	if *peer != "" {
		// Act as client: send ping
		resp, err := http.Get(*peer)
		if err != nil {
			log.Fatalf("Ping failed: %v", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Received response from peer: %s\n", string(body))
		return
	}

	// Act as server: serve ping endpoint
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("🟢 OAIM Agent listening on %s\n", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
