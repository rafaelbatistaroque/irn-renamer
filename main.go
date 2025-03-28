package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Define command-line flags
var (
	oldName = flag.String("old", "", "The old name (current template name) to search for.")
	newName = flag.String("new", "", "The new name (desired microservice name) to replace with.")
)

// List of file extensions to process for content replacement
var processableExtensions = map[string]bool{
	".sln":    true,
	".csproj": true,
	".cs":     true,
	".json":   true,
	// Add other extensions if needed
}

// --- Structs for collecting and reporting results ---

// Struct to hold path information for sorting before rename (Pass 1 -> Pass 2)
type PathInfo struct {
	Path string
	Info fs.FileInfo
}

// Struct to hold info about ONE renamed item
type RenamedItem struct {
	OldPath string
	NewPath string
	// IsDir is implicitly known by the category key ("Directory" vs file extension)
}

// Struct to hold all changes for a specific file type/category
type ChangeReport struct {
	Category       string   // e.g., ".cs", ".csproj", "Dockerfile", "Directory"
	ContentUpdates []string // List of ORIGINAL paths whose content was updated
	Renames        []RenamedItem
}

// Helper to get category key
func getCategoryKey(path string, isDir bool) string {
	if isDir {
		return "Directory"
	}
	ext := filepath.Ext(path)
	baseName := filepath.Base(path)
	if baseName == "Dockerfile" {
		return "Dockerfile"
	} else if ext == "" {
		return "(No Extension)"
	}
	return ext // .cs, .csproj, .sln etc.
}

func main() {
	flag.Parse()

	// --- Validation ---
	if *oldName == "" || *newName == "" {
		fmt.Println("Error: Both -old and -new flags are required.")
		flag.Usage()
		os.Exit(1)
	}
	if *oldName == *newName {
		fmt.Println("Error: -old and -new names cannot be the same.")
		os.Exit(1)
	}

	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}

	fmt.Printf("Starting process in directory: %s\n", rootDir)
	fmt.Printf("Replacing all occurrences of '%s' with '%s'\n", *oldName, *newName)
	fmt.Println("--- IMPORTANT: Make sure you have a backup or are using version control! ---")

	// --- Data structures to collect results ---
	pathsToRename := []PathInfo{}
	// Map key: category string, Value: report struct for that category
	changesByType := make(map[string]*ChangeReport)

	// Helper function to ensure report struct exists for a category
	ensureReportExists := func(categoryKey string) *ChangeReport {
		if _, ok := changesByType[categoryKey]; !ok {
			changesByType[categoryKey] = &ChangeReport{Category: categoryKey}
		}
		return changesByType[categoryKey]
	}

	// --- Pass 1: Update content and collect paths to rename ---
	fmt.Println("\n--- Pass 1: Scanning files and updating content ---")

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		info, infoErr := d.Info() // Get FileInfo once

		// --- Collect paths for potential renaming (files and directories) ---
		if strings.Contains(d.Name(), *oldName) {
			if infoErr != nil {
				fmt.Printf("Warning: Could not get FileInfo for %s, skipping potential rename tracking.\n", path)
			} else {
				pathsToRename = append(pathsToRename, PathInfo{Path: path, Info: info})
			}
		}

		// --- Process file content ---
		if !d.IsDir() {
			isDockerfile := d.Name() == "Dockerfile"
			ext := filepath.Ext(path)
			processThisFile := processableExtensions[ext] || isDockerfile

			if processThisFile {
				contentBytes, readErr := os.ReadFile(path)
				if readErr != nil {
					fmt.Printf("Error reading file %s: %v\n", path, readErr)
					return nil // Continue walking
				}
				content := string(contentBytes)

				if strings.Contains(content, *oldName) {
					newContent := strings.ReplaceAll(content, *oldName, *newName)
					perm := fs.FileMode(0644)
					if infoErr == nil { // Use original permissions if info available
						perm = info.Mode().Perm()
					}
					writeErr := os.WriteFile(path, []byte(newContent), perm)
					if writeErr != nil {
						fmt.Printf("Error writing updated content to file %s: %v\n", path, writeErr)
						return writeErr
					}

					// --- Store content update result under the correct category ---
					categoryKey := getCategoryKey(path, false) // false because we are in !d.IsDir() block
					report := ensureReportExists(categoryKey)
					report.ContentUpdates = append(report.ContentUpdates, path) // Store original path
				}
			}
		}
		return nil // Continue walking
	}

	err = filepath.WalkDir(rootDir, walkFunc)
	if err != nil {
		log.Printf("Warning: Directory walk completed with errors: %v", err)
	}
	fmt.Println("--- Pass 1 Complete ---")

	// --- Pass 2: Rename files and folders ---
	fmt.Println("\n--- Pass 2: Renaming files and directories ---")

	// Sort paths by length descending (basic strategy)
	sort.Slice(pathsToRename, func(i, j int) bool {
		return len(pathsToRename[i].Path) > len(pathsToRename[j].Path)
	})

	renamedCount := 0
	for _, pathInfo := range pathsToRename {
		oldPath := pathInfo.Path

		if _, statErr := os.Stat(oldPath); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			fmt.Printf("Error checking status of %s before rename: %v\n", oldPath, statErr)
			continue
		}

		dir := filepath.Dir(oldPath)
		baseName := filepath.Base(oldPath)

		if strings.Contains(baseName, *oldName) { // Check again, might be redundant but safe
			newBaseName := strings.ReplaceAll(baseName, *oldName, *newName)
			newPath := filepath.Join(dir, newBaseName)

			if _, statErr := os.Stat(newPath); statErr == nil {
				fmt.Printf("Warning: Target path %s already exists. Skipping rename for %s.\n", newPath, oldPath)
				continue
			}

			renameErr := os.Rename(oldPath, newPath)
			if renameErr != nil {
				fmt.Printf("Error renaming %s to %s: %v\n", oldPath, newPath, renameErr)
			} else {
				// --- Store rename result under the correct category ---
				isDir := pathInfo.Info.IsDir()                // Get IsDir from collected PathInfo
				categoryKey := getCategoryKey(oldPath, isDir) // Determine category based on OLD path
				report := ensureReportExists(categoryKey)
				report.Renames = append(report.Renames, RenamedItem{OldPath: oldPath, NewPath: newPath})
				renamedCount++
			}
		}
	}
	fmt.Println("--- Pass 2 Complete ---")

	// --- Final Report ---
	fmt.Println("\n--- Summary of Changes by Type ---")

	// Get and sort categories for consistent report order
	categories := make([]string, 0, len(changesByType))
	for k := range changesByType {
		categories = append(categories, k)
	}
	// Custom sort: Put "Directory" first, then sort others
	sort.SliceStable(categories, func(i, j int) bool {
		if categories[i] == "Directory" {
			return true
		}
		if categories[j] == "Directory" {
			return false
		}
		return categories[i] < categories[j] // Regular sort for the rest
	})

	for _, category := range categories {
		report := changesByType[category]
		hasRenames := len(report.Renames) > 0
		hasContentUpdates := len(report.ContentUpdates) > 0

		if hasRenames || hasContentUpdates {
			fmt.Printf("\n--- Changes for %s ---\n", report.Category)

			if hasRenames {
				fmt.Println("  Renamed:")
				// Sort renames by old path for consistency? Optional.
				sort.Slice(report.Renames, func(i, j int) bool {
					return report.Renames[i].OldPath < report.Renames[j].OldPath
				})
				for _, item := range report.Renames {
					// Show only base name changes for clarity
					fmt.Printf("    - '%s' -> '%s'\n", filepath.Base(item.OldPath), filepath.Base(item.NewPath))
				}
			}

			if hasContentUpdates {
				// Add spacing if renames were also listed
				if hasRenames {
					fmt.Println() // Add a blank line for separation
				}
				fmt.Println("  Content Updated:")
				sort.Strings(report.ContentUpdates) // Sort paths
				for _, path := range report.ContentUpdates {
					relativePath, err := filepath.Rel(rootDir, path)
					if err != nil {
						relativePath = path // Fallback
					}
					fmt.Printf("    - %s\n", relativePath)
				}
			}
		}
	}

	fmt.Printf("\n--- Process Complete ---\n")
	fmt.Printf("Total items renamed: %d\n", renamedCount) // Keep total count for reference
	fmt.Println("Please review the changes carefully and test your project.")
}
