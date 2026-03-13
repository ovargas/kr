package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/ovargas/kr/internal/server"
)

func main() {
	port := flag.Int("port", 0, "port to listen on (0 for random)")
	path := flag.String("path", ".", "path to documentation directory")
	title := flag.String("title", "", "project name shown in navbar (defaults to current directory name)")
	flag.Parse()

	absPath, err := filepath.Abs(*path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving path: %v\n", err)
		os.Exit(1)
	}

	projectName := *title
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
			os.Exit(1)
		}
		projectName = filepath.Base(cwd)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %s is not a directory\n", absPath)
		os.Exit(1)
	}

	srv, err := server.New(*port, absPath, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing server: %v\n", err)
		os.Exit(1)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	case <-sigCh:
		fmt.Println("\nshutting down")
	}
}
