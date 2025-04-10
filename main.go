package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	root := os.Args[1]
	count := 0

	// Регулярка для выделения версии
	versionRegex := regexp.MustCompile(`(v[\d]+(?:[\.\-\w]*)?)$`)

	// Настройка табличного вывода
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Index Name\tVersion")

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			return nil
		}

		if d.IsDir() && d.Name() == "target" {
			return filepath.SkipDir
		}

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
				version := "-"
				match := versionRegex.FindStringSubmatch(indexName)
				if len(match) > 1 {
					version = match[1]
				}
				fmt.Fprintf(writer, "%s\t%s\n", indexName, version)
				count++
			} else {
				fmt.Fprintf(os.Stderr, "%s: 'indexName' not found or not a string\n", path)
			}
		}
		return nil
	})

	writer.Flush()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTotal indexName entries found: %d\n", count)
}
