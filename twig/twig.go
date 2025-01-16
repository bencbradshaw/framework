package twig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Template struct {
	RawContent string
	Blocks     map[string]string
}

var templatesDir = "./templates"

func Render(templatePath string, data map[string]interface{}) (string, error) {
	fullTemplatePath := filepath.Join(templatesDir, templatePath)
	fmt.Println("fullTemplatePath: ", fullTemplatePath)
	// Phase 1: Analyze Templates
	templateAnalysis, err := analyzeTemplates(fullTemplatePath)
	if err != nil {
		return "", err
	}

	// Phase 2: Build Template
	builtTemplate, err := buildTemplate(templateAnalysis, fullTemplatePath, data)
	if err != nil {
		return "", err
	}

	// Phase 3: Clean and Check Template
	finalTemplate, err := cleanAndCheck(builtTemplate)
	if err != nil {
		return "", err
	}

	return finalTemplate, nil
}

type TemplateInfo struct {
	Path          string
	Extends       string
	Includes      []string
	DefinedBlocks map[string]string // Where we store block definitions
	UsedBlocks    map[string]string // Where blocks are used
	EntryTemplate string            // Path to the original entry template
}

func analyzeTemplates(templatePath string) (map[string]TemplateInfo, error) {
	templateAnalysis := make(map[string]TemplateInfo)

	// Define a recursive function to analyze each template file
	var analyzeTemplate func(path string, isEntry bool) error

	analyzeTemplate = func(path string, isEntry bool) error {
		fmt.Printf("Analyzing template at path: %s\n", path)
		if _, exists := templateAnalysis[path]; exists {
			// Avoid reanalyzing a template that has already been processed
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}

		content := string(raw)
		extends, includes, definedBlocks, usedBlocks := parseTemplate(content)

		// Create a new TemplateInfo instance
		templateInfo := TemplateInfo{
			Path:          path,
			Extends:       extends,
			Includes:      includes,
			DefinedBlocks: definedBlocks,
			UsedBlocks:    usedBlocks,
		}

		// If this is the entry template, set the EntryTemplate field
		if isEntry {
			templateInfo.EntryTemplate = path
		}

		// Store the TemplateInfo struct back into the map
		templateAnalysis[path] = templateInfo

		// Recursively analyze included templates and parent templates
		for _, include := range includes {
			includePath := filepath.Join(templatesDir, include)
			err := analyzeTemplate(includePath, false) // Entry marker false for includes
			if err != nil {
				return err
			}
		}

		if extends != "" {
			parentPath := filepath.Join(templatesDir, extends)
			err := analyzeTemplate(parentPath, false) // Entry marker false for extends
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Start the analysis from the full template path
	err := analyzeTemplate(templatePath, true) // Mark the initial template as the entry
	if err != nil {
		return nil, err
	}
	fmt.Printf("templateAnalysis: %+v\n", templateAnalysis)
	return templateAnalysis, nil
}

func parseTemplate(content string) (string, []string, map[string]string, map[string]string) {
	var extends string
	var includes []string
	definedBlocks := make(map[string]string)
	usedBlocks := make(map[string]string)

	// Regex for detecting extends
	extendsRegex := regexp.MustCompile(`{% extends "([^"]+)" %}`)
	if matches := extendsRegex.FindStringSubmatch(content); len(matches) > 1 {
		extends = matches[1]
	}

	// Regex for detecting includes
	includeRegex := regexp.MustCompile(`{% include "([^"]+)" %}`)
	includeMatches := includeRegex.FindAllStringSubmatch(content, -1)
	for _, match := range includeMatches {
		if len(match) > 1 {
			includes = append(includes, match[1])
		}
	}

	// Regex for detecting block definitions and their usages
	blockDefRegex := regexp.MustCompile(`(?s){% block (\w+) %}(.*?){% endblock %}`)
	blockMatches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range blockMatches {
		if len(match) > 2 {
			blockName := match[1]
			definedBlocks[blockName] = strings.TrimSpace(match[2]) // Store the definition
		}
	}

	// Regex for detecting block usages
	blockUsageRegex := regexp.MustCompile(`{{\s*block\("(\w+)"\)\s*}}`)
	usageMatches := blockUsageRegex.FindAllStringSubmatch(content, -1)
	for _, match := range usageMatches {
		if len(match) > 1 {
			blockName := match[1]
			usedBlocks[blockName] = match[0] // Keep the original usage for later replacement
		}
	}

	return extends, includes, definedBlocks, usedBlocks
}

func buildTemplate(analysis map[string]TemplateInfo, templatePath string, data map[string]interface{}) (string, error) {
	var finalContent strings.Builder

	var renderTemplate func(path string, parentBlocks map[string]string) (string, map[string]string, error)

	renderTemplate = func(path string, parentBlocks map[string]string) (string, map[string]string, error) {
		templateInfo, exists := analysis[path]
		if !exists {
			return "", nil, fmt.Errorf("template not found: %s", path)
		}

		// Replace includes first
		content := replaceIncludes(templateInfo.Path, templateInfo, analysis)

		// Handle block usages before we check for inheritance
		for blockName, usage := range templateInfo.UsedBlocks {
			if definedContent, exists := templateInfo.DefinedBlocks[blockName]; exists {
				content = strings.ReplaceAll(content, usage, definedContent)
			}
		}

		if templateInfo.Extends != "" {
			// Render the parent template first
			parentContent, parentBlocks, err := renderTemplate(filepath.Join(templatesDir, templateInfo.Extends), nil)
			if err != nil {
				return "", nil, err
			}

			// Merge the current content with the parent's content
			mergedContent := mergeBlocks(parentContent, parentBlocks, templateInfo.DefinedBlocks)
			return mergedContent, templateInfo.DefinedBlocks, nil
		} else {
			// In the base case, merge the final content with any parent blocks
			completeContent := mergeBlocks(content, parentBlocks, templateInfo.DefinedBlocks)
			return completeContent, templateInfo.DefinedBlocks, nil
		}
	}

	// Start rendering from the main template path
	content, _, err := renderTemplate(templatePath, nil)
	if err != nil {
		return "", err
	}

	// Handle variable replacements in the content
	content = handleVars(content, data)

	// Final pass to replace blocks using the EntryTemplate
	entryTemplateInfo := analysis[templatePath]
	if entryTemplateInfo.EntryTemplate != "" {
		// Replace all block placeholders based on the entry template
		for blockName, definedContent := range entryTemplateInfo.DefinedBlocks {
			placeholder := fmt.Sprintf("{%% block %s %%}{%% endblock %%}", blockName)
			content = strings.ReplaceAll(content, placeholder, definedContent)
		}
	}

	finalContent.WriteString(content)
	return finalContent.String(), nil
}

func replaceIncludes(templatePath string, tInfo TemplateInfo, analysis map[string]TemplateInfo) string {
	fmt.Println("Replacing includes in template:", templatePath)
	fmt.Println("Includes:", tInfo.Includes)
	content := readFileContent(templatePath)

	// Replace includes with actual content
	for _, include := range tInfo.Includes {
		fmt.Println("Processing include", include)
		includePath := filepath.Join(filepath.Dir(templatePath), include)
		includedContent := readFileContent(includePath)
		fmt.Println("Included content", includedContent)
		includeTag := `{%` + ` include ` + `"` + include + `"` + ` %}`
		fmt.Println("include tag", includeTag)
		content = strings.Replace(content, includeTag, includedContent, 1)
		fmt.Println("Content after include", content)
	}
	return content
}

func readFileContent(path string) string {
	fmt.Println("Reading file at path", path)
	raw, _ := os.ReadFile(path)
	return string(raw)
}

func mergeBlocks(content string, parentBlocks, childBlocks map[string]string) string {
	blockDefRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)

	// Iterate over all block definitions in the content
	matches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		// Assuming match[2] contains the block type
		blockType := match[2]
		if blockType != "UsedBlock" {
			continue
		}

		blockName := match[1]
		blockPlaceholder := match[0]
		println("blockName: ", blockName)
		println("blockPlaceholder: ", blockPlaceholder)
		if childContent, exists := childBlocks[blockName]; exists {
			// Replace the entire block definition with the child's content
			println("block child content ", childContent)
			content = strings.Replace(content, blockPlaceholder, childContent, 1)
		} else if parentContent, exists := parentBlocks[blockName]; exists {
			// If no child block, use parent's block content
			println("block parent content ", parentContent)
			content = strings.Replace(content, blockPlaceholder, parentContent, 1)
		}
	}
	println("RETURN BLOCK ", content)
	return content
}

func handleVars(content string, data map[string]interface{}) string {
	fmt.Println("HANDLEVARS. content")
	for key, value := range data {
		placeholder := "{{ " + key + " }}"
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
		fmt.Println("Replacing variable %s with %v\n", key, value)
	}
	return content
}

func cleanAndCheck(content string) (string, error) {
	// Step 1: Ensure completeness by checking for any unresolved blocks
	unresolvedBlockRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	unresolvedBlocks := unresolvedBlockRegex.FindAllStringSubmatch(content, -1)

	if len(unresolvedBlocks) > 0 {
		fmt.Println("unresolved blocks found in template: %v", unresolvedBlocks)
		// now replace all of them with empty strings
		for _, block := range unresolvedBlocks {
			content = strings.Replace(content, block[0], "", -1)
		}
	}

	// Step 2: (Optional) Final cleanup operations can be added here, like ensuring newline consistency, trimming, etc.
	// This is heavily dependent on the specific requirements of your application

	// Step 3: Validate HTML or specific syntax if needed. Here it's assumed to be a simple pass-through.
	// Add syntax checking logic if there is a specific requirement that is not covered by previous steps

	// Log the cleaned content if needed for debugging
	log.Println("Final Cleaned Content:")
	log.Println(content)

	return content, nil
}
