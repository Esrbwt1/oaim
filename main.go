package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: oaim <command>")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "init":
        cmdInit()
    default:
        fmt.Printf("Unknown command: %s\n", os.Args[1])
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
    description: "Micropayments & revenue reporting"
`
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
