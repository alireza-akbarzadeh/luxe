package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Get module name from go.mod
	modBytes, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Println("Error reading go.mod:", err)
		return
	}
	modLine := strings.TrimPrefix(strings.SplitN(string(modBytes), "\n", 1)[0], "module ")
	moduleName := strings.TrimSpace(modLine)

	// Regex to match Swagger comments that reference models.Type
	re := regexp.MustCompile(`(//\s*@(Success|Failure|Param|Router|Response).*\{object\})\s+models\.([A-Z][a-zA-Z0-9_]*)`)

	err = filepath.Walk("internal/interfaces/http/handlers", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		newContent := re.ReplaceAllFunc(content, func(match []byte) []byte {
			// match is e.g. "// @Success 200 {object} models.Order"
			// Replace "models.Order" with "github.com/yourmodule/internal/models.Order"
			subs := re.FindSubmatch(match)
			if len(subs) != 4 {
				return match
			}
			prefix := subs[1] // "// @Success 200 {object}"
			typeName := subs[3]
			replacement := fmt.Sprintf("%s %s/internal/models.%s", prefix, moduleName, typeName)
			return []byte(replacement)
		})
		if !bytes.Equal(content, newContent) {
			if err := os.WriteFile(path, newContent, 0644); err != nil {
				return err
			}
			fmt.Printf("Updated: %s\n", path)
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking files:", err)
	}
}
