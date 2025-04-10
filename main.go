package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			return nil
		}

		// Проверяем, что файл .json и в пути есть "read-view"
		if !d.IsDir() && strings.Contains(path, "read-view") && strings.HasSuffix(path, ".json") {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to read %s: %v\n", path, err)
				return nil
			}

			var jsonData map[string]interface{}
			if err := json.Unmarshal(data, &jsonData); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to parse JSON in %s: %v\n", path, err)
				return nil
			}

			if indexName, ok := jsonData["indexName"].(string); ok {
				fmt.Printf("%s\n", indexName)
			} else {
				fmt.Fprintf(os.Stderr, "%s: 'indexName' not found or not a string\n", path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk error: %v\n", err)
		os.Exit(1)
	}
}
