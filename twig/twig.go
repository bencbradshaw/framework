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
	log.Printf("fullTemplatePath: ", fullTemplatePath)
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
	Path     string
	Extends  string
	Includes []string
	Blocks   map[string]string
}

func analyzeTemplates(templatePath string) (map[string]TemplateInfo, error) {

	templateAnalysis := make(map[string]TemplateInfo)

	// Define a recursive function to analyze each template file
	var analyzeTemplate func(path string) error

	analyzeTemplate = func(path string) error {
		log.Printf("Analyzing template at path: %s\n", path)
		if _, exists := templateAnalysis[path]; exists {
			// Avoid reanalyzing a template that has already been processed
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}

		content := string(raw)
		extends, includes, blocks := parseTemplate(content)

		templateInfo := TemplateInfo{
			Path:     path,
			Extends:  extends,
			Includes: includes,
			Blocks:   blocks,
		}

		templateAnalysis[path] = templateInfo

		// Recursively analyze included templates and parent templates
		for _, include := range includes {
			includePath := filepath.Join(templatesDir, include)
			err := analyzeTemplate(includePath)
			if err != nil {
				return err
			}
		}

		if extends != "" {
			parentPath := filepath.Join(templatesDir, extends)
			err := analyzeTemplate(parentPath)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Start the analysis from the full template path
	err := analyzeTemplate(templatePath)
	if err != nil {
		return nil, err
	}
	fmt.Printf("templateAnalysis: %+v\n", templateAnalysis)
	return templateAnalysis, nil
}

func parseTemplate(content string) (string, []string, map[string]string) {
	var extends string
	var includes []string
	blocks := make(map[string]string)

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

	// Regex for detecting blocks
	blockDefRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	blockMatches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range blockMatches {
		if len(match) > 2 {
			blocks[match[1]] = match[2]
		}
	}

	return extends, includes, blocks
}

func buildTemplate(analysis map[string]TemplateInfo, templatePath string, data map[string]interface{}) (string, error) {
	var finalContent strings.Builder

	// Helper function to render a specific template
	var renderTemplate func(path string, parentBlocks map[string]string) (string, map[string]string, error)

	renderTemplate = func(path string, parentBlocks map[string]string) (string, map[string]string, error) {

		templateInfo, exists := analysis[path]
		if strings.Contains(analysis[path].Path, "templates/") {
			templateInfo, exists = analysis[path]
		} else {
			templateInfo, exists = analysis["templates/"+path]
		}
		if !exists {
			return "", nil, fmt.Errorf("template not found: %s", path)
		}

		if templateInfo.Extends != "" { // Template extends a parent template
			// First render the parent template to get its content and blocks
			parentContent, parentBlocks, err := renderTemplate(templateInfo.Extends, nil)
			if err != nil {
				return "", nil, err
			}

			// Merge blocks from this template (child) with the parent's blocks
			mergedContent := mergeBlocks(parentContent, parentBlocks, templateInfo.Blocks)
			return mergedContent, templateInfo.Blocks, nil
		} else {
			// Base case: no parent, render this template directly
			content := replaceIncludes(templateInfo.Path, templateInfo, analysis)

			completeContent := mergeBlocks(content, parentBlocks, templateInfo.Blocks)
			return completeContent, templateInfo.Blocks, nil
		}
	}

	// Start rendering from the main template path
	content, _, err := renderTemplate(templatePath, nil)
	if err != nil {
		return "", err
	}

	// Replace variables in the final content
	content = handleVars(content, data)

	finalContent.WriteString(content)
	return finalContent.String(), nil
}

func replaceIncludes(templatePath string, tInfo TemplateInfo, analysis map[string]TemplateInfo) string {
	content := readFileContent(templatePath)

	// Replace includes with actual content
	for _, include := range tInfo.Includes {
		includePath := filepath.Join(filepath.Dir(templatePath), include)
		includedContent := readFileContent(includePath)
		includeTag := fmt.Sprintf(`{% include "%s" %}`, include)
		content = strings.Replace(content, includeTag, includedContent, 1)
	}
	return content
}

func readFileContent(path string) string {
	log.Printf("Reading file at path: %s\n", path)
	raw, _ := os.ReadFile(path)
	return string(raw)
}

func mergeBlocks(content string, parentBlocks, childBlocks map[string]string) string {
	blockDefRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)

	// Iterate over all block definitions in the content
	matches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		blockName := match[1]
		blockPlaceholder := match[0]

		if childContent, exists := childBlocks[blockName]; exists {
			// Replace the entire block definition with the child's content
			content = strings.Replace(content, blockPlaceholder, childContent, 1)
		} else if parentContent, exists := parentBlocks[blockName]; exists {
			// If no child block, use parent's block content
			content = strings.Replace(content, blockPlaceholder, parentContent, 1)
		}
	}
	return content
}

func handleVars(content string, data map[string]interface{}) string {
	log.Printf("HANDLEVARS. content")
	for key, value := range data {
		placeholder := "{{ " + key + " }}"
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
		log.Printf("Replacing variable %s with %v\n", key, value)
	}
	return content
}

func cleanAndCheck(content string) (string, error) {
	// Step 1: Ensure completeness by checking for any unresolved blocks
	unresolvedBlockRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	unresolvedBlocks := unresolvedBlockRegex.FindAllStringSubmatch(content, -1)

	if len(unresolvedBlocks) > 0 {
		return "", fmt.Errorf("unresolved blocks found in template: %v", unresolvedBlocks)
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
