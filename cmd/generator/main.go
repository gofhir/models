package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gofhir/fhir/internal/codegen/generator"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <version>")
	}

	version := os.Args[1]
	if version != "r4" && version != "r4b" && version != "r5" {
		log.Fatal("Version must be r4, r4b, or r5")
	}

	// Get project root (two directories up from cmd/generator)
	execPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're in cmd/generator, go up two levels
	root := execPath
	if filepath.Base(execPath) == "generator" {
		root = filepath.Dir(filepath.Dir(execPath))
	}

	config := generator.Config{
		SpecsDir:    filepath.Join(root, "specs"),
		OutputDir:   filepath.Join(root, version),
		PackageName: version,
		Version:     version,
	}

	log.Printf("Generating %s code...", version)
	log.Printf("Root: %s", root)
	log.Printf("Specs: %s", config.SpecsDir)
	log.Printf("Output: %s", config.OutputDir)

	gen := generator.New(config)
	if err := gen.LoadTypes(); err != nil {
		log.Fatalf("Failed to load types: %v", err)
	}
	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate: %v", err)
	}

	log.Printf("Successfully generated %s", version)
}
