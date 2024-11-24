// test_env.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	// Print current directory
	pwd, _ := os.Getwd()
	fmt.Printf("Current directory: %s\n", pwd)

	// List all files in current directory
	files, _ := filepath.Glob("*")
	fmt.Println("\nFiles in current directory:")
	for _, f := range files {
		info, _ := os.Stat(f)
		fmt.Printf("%s (size: %d bytes)\n", f, info.Size())
	}

	// Try to load .env
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("\nError loading .env: %v\n", err)
	} else {
		fmt.Println("\nSuccessfully loaded .env")
	}

	// Print environment variables
	fmt.Println("\nEnvironment variables:")
	fmt.Printf("TURSO_DATABASE_URL set: %v\n", os.Getenv("TURSO_DATABASE_URL") != "")
	fmt.Printf("TURSO_AUTH_TOKEN set: %v\n", os.Getenv("TURSO_AUTH_TOKEN") != "")
	fmt.Printf("JWT_SECRET set: %v\n", os.Getenv("JWT_SECRET") != "")
}
