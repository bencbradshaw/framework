package twig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Template represents a parsed Twig-like template.
type Template struct {
	RawContent string
	Blocks     map[string]string
}

func Render2(templatePath string, data map[string]interface{}) (string, error) {
	println("RENDER. TEMPLATE: %s\n", templatePath)
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return "", err
	}
	content := string(raw)
	content = handleExtends(content, data)
	content = handleIncludes(content, data)
	content = handleBlocks(content, data)
	content = handleVars(content, data)

	println("returning while in", templatePath)
	return content, nil
}

func handleBlocks(content string, data map[string]interface{}) string {
	println("HANDLEBLOCKS. content: %s\n", content)
	blockDefRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	matches := blockDefRegex.FindAllStringSubmatch(content, -1)
	blocks := make(map[string]string)
	if len(matches) > 0 {
		for _, match := range matches {
			if match[2] != "" {
				blockName := match[1]
				blockContent := match[2]
				println("BlockName", blockName)
				println("BlockContent", blockContent)
				blocks[blockName] = blockContent
			}
		}
	}
	return content
}

func handleExtends(content string, data map[string]interface{}) string {
	println("HANDLEEXTENDS. content: %s\n", content)
	extendsMatch := regexp.MustCompile(`{% extends "([^"]+)" %}`)
	match := extendsMatch.FindStringSubmatch(content)
	var extendsTemplateName string
	if len(match) > 1 {
		extendsTemplateName = match[1]
	}
	if extendsTemplateName != "" {
		println("Matched extends", extendsTemplateName)
		baseTemplatePath := filepath.Join("templates", extendsTemplateName)
		baseOutput, err := Render2(baseTemplatePath, data)
		if err != nil {
			return baseOutput
		}
		return baseOutput
	} else {
		println("No Extends", extendsTemplateName)
		return content
	}
}

func handleIncludes(content string, data map[string]interface{}) string {
	println("HANDLE INCLUDES", content)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		fmt.Println("LINE", line)
		if strings.Contains(strings.TrimSpace(line), "{% include ") {
			fmt.Println("HAS INCLUDE", line)
			// Extract template name
			start := strings.Index(line, `"`) + 1
			end := strings.LastIndex(line, `"`)
			includePath := line[start:end]

			log.Printf("Including template: %s\n", includePath)

			// Construct the full path to the base template
			baseTemplatePath := filepath.Join("templates", includePath)
			// Read and replace the include
			includedContent, err := Render2(baseTemplatePath, data) // Pass the required data here
			if err == nil {
				content = strings.Replace(content, line, includedContent, 1)
				log.Printf("Included template: %s successfully.\n", includePath)
			} else {
				log.Printf("Error including template %s: %v\n", includePath, err)
			}
		}
	}
	return content
}

func handleVars(content string, data map[string]interface{}) string {
	println("HANDLEVARS. content: %s\n", content)
	for key, value := range data {
		placeholder := "{{ " + key + " }}"
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
		log.Printf("Replacing variable %s with %v\n", key, value)
	}
	return content
}
