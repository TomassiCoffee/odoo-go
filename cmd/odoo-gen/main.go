package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	odoo "github.com/TomassiCoffee/odoo-go"
	"github.com/TomassiCoffee/odoo-go/internal/typegen"
)

func main() {
	var outputDir string
	var packageName string
	var odooImportPath string
	var cacheFile string
	var dotenv string
	var modelsCSV string
	var pageSize int

	flag.StringVar(&outputDir, "output-dir", "odoomodels/", "generated Go output directory")
	flag.StringVar(&packageName, "package", "odoomodels", "generated Go package name")
	flag.StringVar(&odooImportPath, "odoo-import", "github.com/TomassiCoffee/odoo-go", "import path used by generated code")
	flag.StringVar(&cacheFile, "cache", "odoomodels/models_metadata.json", "metadata cache file used as fallback")
	flag.StringVar(&dotenv, "env", ".env", "optional .env file")
	flag.StringVar(&modelsCSV, "models", "", "optional comma-separated Odoo model names; empty generates every model")
	flag.IntVar(&pageSize, "page-size", 500, "Odoo metadata page size")
	flag.Parse()

	if err := odoo.LoadDotEnv(dotenv); err != nil {
		fatal(err)
	}
	modelNames := parseModelNames(modelsCSV)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var cache typegen.MetadataCache
	client, err := odoo.NewClientFromEnv()
	if err != nil {
		fallback, fallbackErr := loadCache(cacheFile)
		if fallbackErr != nil {
			fatal(fmt.Errorf("create Odoo client failed: %w; cache fallback failed: %w", err, fallbackErr))
		}
		cache = typegen.FilterCache(fallback, modelNames)
		cache.Source = "cache fallback"
		fmt.Fprintf(os.Stderr, "warning: could not create Odoo client: %v\nusing cache: %s\n", err, cacheFile)
	} else {
		cache, err = typegen.Fetch(ctx, client, pageSize, modelNames)
		if err != nil {
			fallback, fallbackErr := loadCache(cacheFile)
			if fallbackErr != nil {
				fatal(fmt.Errorf("fetch Odoo metadata failed: %w; cache fallback failed: %w", err, fallbackErr))
			}
			cache = typegen.FilterCache(fallback, modelNames)
			cache.Source = "cache fallback"
			fmt.Fprintf(os.Stderr, "warning: could not fetch Odoo metadata: %v\nusing cache: %s\n", err, cacheFile)
		} else if err := saveCache(cacheFile, cache); err != nil {
			fatal(err)
		}
	}

	if len(modelNames) > 0 && len(cache.Models) != len(modelNames) {
		fmt.Fprintf(os.Stderr, "warning: requested %d models, generated metadata for %d\n", len(modelNames), len(cache.Models))
	}

	fmt.Printf("Generating using %d models and %d fields.\n", len(cache.Models), cache.FieldCount)
	sources, err := typegen.Render(cache, typegen.RenderOptions{
		PackageName:    packageName,
		OdooImportPath: odooImportPath,
	})
	if err != nil {
		fatal(err)
	}

	for name, source := range sources{
		if err := os.MkdirAll(filepath.Dir(outputDir), 0o755); err != nil {
			fatal(err)
		}
		output := filepath.Join(outputDir, name + ".gen.go")
		fmt.Printf("Writing output %s from %s.\n", output, cache.Source)
		if err := os.WriteFile(output, source, 0o644); err != nil {
			fatal(err)
		}
	}
}

func parseModelNames(csv string) []string {
	if strings.TrimSpace(csv) == "" {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, part := range strings.Split(csv, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func loadCache(path string) (typegen.MetadataCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return typegen.MetadataCache{}, err
	}
	var cache typegen.MetadataCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return typegen.MetadataCache{}, err
	}
	return cache, nil
}

func saveCache(path string, cache typegen.MetadataCache) error {
	data, err := typegen.SaveCacheJSON(cache)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
