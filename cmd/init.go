package cmd

import (
	"bufio"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed templates
var templatesFS embed.FS

var (
	templateName string
	forceFlag    bool
	outputPath   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize basic tasks.json file",
	Long:  `Initialize a basic tasks.json file with template support.`,
	RunE:  runInitCommand,
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&templateName, "template", "t", "default", "template to use (default, go, node)")
	initCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "overwrite existing file")
	initCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output path (default: .vscode/tasks.json)")
}

func runInitCommand(cmd *cobra.Command, args []string) error {
	// Determine output path
	targetPath := outputPath
	if targetPath == "" {
		targetPath = ".vscode/tasks.json"
	}

	// Check if file already exists
	if !forceFlag {
		if _, err := os.Stat(targetPath); err == nil {
			if !confirmOverwrite(targetPath) {
				fmt.Println("Operation cancelled.")
				return nil
			}
		}
	}

	// Validate template
	if !isValidTemplate(templateName) {
		return fmt.Errorf("invalid template '%s'. Available templates: default, go, node", templateName)
	}

	// Create directory if it doesn't exist
	if err := createDirectoryIfNeeded(targetPath); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Load template content
	templateContent, err := getTemplateContent(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Write template to file
	if err := writeTemplateToFile(targetPath, templateContent); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	if !quiet {
		fmt.Printf("Successfully created %s using '%s' template\n", targetPath, templateName)
		if verbose {
			fmt.Printf("Template path: templates/%s.json\n", templateName)
		}
	}

	return nil
}

func isValidTemplate(template string) bool {
	validTemplates := []string{"default", "go", "node"}
	for _, valid := range validTemplates {
		if template == valid {
			return true
		}
	}
	return false
}

func confirmOverwrite(path string) bool {
	if quiet {
		return false
	}

	fmt.Printf("File %s already exists. Overwrite? (y/N): ", path)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func createDirectoryIfNeeded(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir == "." {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

func getTemplateContent(templateName string) (string, error) {
	templatePath := fmt.Sprintf("templates/%s.json", templateName)
	content, err := templatesFS.ReadFile(templatePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeTemplateToFile(path string, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(content)
	return err
}

func GetAvailableTemplates() []string {
	return []string{"default", "go", "node"}
}

func GetTemplateContent(templateName string) (string, error) {
	return getTemplateContent(templateName)
}