package main

import (
	"bufio"
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
	oldName = flag.String("old", "", "Antigo valor/nome que pretende substituir")
	newName = flag.String("new", "", "Novo nome/valor que pretende usar")
)

// List of file extensions to process for content replacement
var processableExtensions = map[string]bool{
	".sln":    true,
	".csproj": true,
	".cs":     true,
	".json":   true,
}

var ignoredDirs = map[string]bool{
	".git":         true,
	".vscode":      true,
	".idea":        true,
	"bin":          true,
	"obj":          true,
	"node_modules": true,
}

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
		fmt.Println("ERRO: Os parámetros -old e -new são obrigatórios.")
		flag.Usage()
		os.Exit(1)
	}
	if *oldName == *newName {
		fmt.Println("ERRO: os valores de -old e -new são os mesmos.")
		os.Exit(1)
	}

	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Erro ao obter diretório atual: %v", err)
	}

	fmt.Printf("Iniciando processo no diretório: %s\n", rootDir)
	fmt.Printf("Substituir ocorrências de '%s' para '%s'\n", *oldName, *newName)

	// --- !! BLOCO DE CONFIRMAÇÃO ADICIONADO !! ---
	fmt.Println("\n------------------------------------ ATENÇÃO ------------------------------------")
	fmt.Println("Este processo modificará ficheiros, nomes de pastas e conteúdo DIRETAMENTE.")
	fmt.Println("Garanta que você possui um BACKUP ou está a usar um controle de versão (Git)")
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Print("Você tem certeza que deseja continuar? (Digite S ou Y para confirmar): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Erro ao ler a confirmação do usuário: %v", err)
	}

	// Limpa espaços em branco (incluindo \n) e converte para maiúsculas
	confirmation := strings.ToUpper(strings.TrimSpace(input))

	if confirmation != "S" && confirmation != "Y" {
		fmt.Println("\nProcesso abortado pelo usuário.")
		os.Exit(0)
	}

	pathsToRename := []PathInfo{}
	changesByType := make(map[string]*ChangeReport)

	ensureReportExists := func(categoryKey string) *ChangeReport {
		if _, ok := changesByType[categoryKey]; !ok {
			changesByType[categoryKey] = &ChangeReport{Category: categoryKey}
		}
		return changesByType[categoryKey]
	}

	// --- Pass 1: Update content and collect paths to rename ---
	fmt.Println("\n--- Passo 1: Analizar ficheiros e substituir conteúdo ---")
	renamedCount := 0
	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Erro ao acessar caminho %q: %v\n", path, err)
			return err
		}

		// --- !! VERIFICAÇÃO PARA IGNORAR DIRETÓRIOS ADICIONADA !! ---
		if d.IsDir() { // Verifica se é um diretório
			dirName := d.Name()
			if ignoredDirs[dirName] { // Verifica se o nome está no mapa de ignorados
				// fmt.Printf("Skipping directory: %s\n", path) // Mensagem opcional de debug/info
				return filepath.SkipDir // Instrução para NÃO entrar neste diretório
			}
		}

		info, infoErr := d.Info() // Get FileInfo once

		// --- Collect paths for potential renaming (files and directories) ---
		if strings.Contains(d.Name(), *oldName) {
			if infoErr != nil {
				fmt.Printf("ATENÇÃO: Não foi possível obter FileInfo para %s, ignorando análise e renomeação.\n", path)
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
					fmt.Printf("Erro ao ler ficheiro %s: %v\n", path, readErr)
					return nil // Continue walking
				}
				content := string(contentBytes)

				if strings.Contains(content, *oldName) {
					newContent := strings.ReplaceAll(content, *oldName, *newName)
					perm := fs.FileMode(0644)
					renamedCount++
					if infoErr == nil { // Use original permissions if info available
						perm = info.Mode().Perm()
					}
					writeErr := os.WriteFile(path, []byte(newContent), perm)
					if writeErr != nil {
						fmt.Printf("Erro ao atualizar conteúdo do ficheiro %s: %v\n", path, writeErr)
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
		log.Printf("ATENÇÃO: Análise de diretório com erros: %v", err)
	}
	fmt.Println("--- Passo 1: Completo ---")

	// --- Pass 2: Rename files and folders ---
	fmt.Println("\n--- Passo 2: Renomear ficheiros e diretórios ---")

	// Sort paths by length descending (basic strategy)
	sort.Slice(pathsToRename, func(i, j int) bool {
		return len(pathsToRename[i].Path) > len(pathsToRename[j].Path)
	})

	for _, pathInfo := range pathsToRename {
		oldPath := pathInfo.Path

		if _, statErr := os.Stat(oldPath); os.IsNotExist(statErr) {
			continue
		} else if statErr != nil {
			fmt.Printf("Erro ao verificar status de %s antes de renomear: %v\n", oldPath, statErr)
			continue
		}

		dir := filepath.Dir(oldPath)
		baseName := filepath.Base(oldPath)

		if strings.Contains(baseName, *oldName) { // Check again, might be redundant but safe
			newBaseName := strings.ReplaceAll(baseName, *oldName, *newName)
			newPath := filepath.Join(dir, newBaseName)

			if _, statErr := os.Stat(newPath); statErr == nil {
				fmt.Printf("ATENÇÃO: Diretório %s já existe. Ignorar renomeação para %s.\n", newPath, oldPath)
				continue
			}

			renameErr := os.Rename(oldPath, newPath)
			if renameErr != nil {
				fmt.Printf("Erro ao renomear %s para %s: %v\n", oldPath, newPath, renameErr)
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
	fmt.Println("--- Passo 2: Completo ---")

	// --- Final Report ---
	fmt.Print("\n--- Resumo por tipos:")

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
			fmt.Printf("\n--- Tipo %s ---\n", report.Category)

			if hasRenames {
				fmt.Println("  Renomeado:")
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
				fmt.Println("  Conteúdo Atualizado:")
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

	fmt.Printf("\n--- Processo Finalizado ---\n")
	fmt.Printf("Total de itens renomeados: %d\n", renamedCount)
	fmt.Println("Por favor, revise as alterações com cuidado e teste seu projeto.")
}
