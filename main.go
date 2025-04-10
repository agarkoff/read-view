package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"
)

func isInsideTarget(path string) bool {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		if part == "target" {
			return true
		}
	}
	return false
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <directory> <elasticsearch_url>")
		os.Exit(1)
	}

	root := os.Args[1]
	esURL := strings.TrimRight(os.Args[2], "/")
	count := 0

	versionRegex := regexp.MustCompile(`(v[\d]+(?:[\.\-\w]*)?)$`)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "Index Name\tVersion\tExists")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	seen := make(map[string]string) // кэш: indexName -> exists ("yes"/"no")

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accessing %s: %v\n", path, err)
			return nil
		}

		if isInsideTarget(path) {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".json") || !strings.Contains(path, "read-view") {
			return nil
		}

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

		indexName, ok := jsonData["indexName"].(string)
		if !ok {
			return nil
		}

		version := "-"
		match := versionRegex.FindStringSubmatch(indexName)
		if len(match) > 1 {
			version = match[1]
		}

		exists, ok := seen[indexName]
		if !ok {
			req, err := http.NewRequest("HEAD", fmt.Sprintf("%s/%s", esURL, indexName), nil)
			if err == nil {
				resp, err := client.Do(req)
				if err == nil {
					if resp.StatusCode == 200 {
						exists = "yes"
					} else if resp.StatusCode == 404 {
						exists = "no"
					} else {
						exists = "unknown"
					}
					resp.Body.Close()
				} else {
					exists = "unknown"
				}
			} else {
				exists = "unknown"
			}
			seen[indexName] = exists
		}

		fmt.Fprintf(writer, "%s\t%s\t%s\n", indexName, version, exists)
		count++
		return nil
	})

	writer.Flush()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Walk error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTotal indexName entries found: %d\n", count)
}
