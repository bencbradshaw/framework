package templating

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
)

var templatesDir = "./templates"

func HtmlRender(templatePath string, data map[string]interface{}) (string, error) {
	fullTemplatePath := filepath.Join(templatesDir, templatePath)
	templateAnalysis, err := analyzeTemplates(fullTemplatePath)
	if err != nil {
		return "", err
	}

	finalContent, err := buildTemplate(templateAnalysis, fullTemplatePath, data)
	if err != nil {
		return "", err
	}

	return finalContent, nil
}

type TemplateInfo struct {
	Path          string
	Extends       string
	Includes      []string
	DefinedBlocks map[string]string
	UsedBlocks    map[string]string
}

func analyzeTemplates(templatePath string) (map[string]TemplateInfo, error) {
	templateAnalysis := make(map[string]TemplateInfo)
	var analyzeTemplate func(path string) error
	analyzeTemplate = func(path string) error {
		if _, exists := templateAnalysis[path]; exists {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}
		content := string(raw)
		extends, includes, definedBlocks, usedBlocks := parseTemplate(content)
		templateInfo := TemplateInfo{
			Path:          path,
			Extends:       extends,
			Includes:      includes,
			DefinedBlocks: definedBlocks,
			UsedBlocks:    usedBlocks,
		}
		templateAnalysis[path] = templateInfo

		for _, include := range includes {
			includePath := filepath.Join(templatesDir, include)
			if err := analyzeTemplate(includePath); err != nil {
				return err
			}
		}
		if extends != "" {
			parentPath := filepath.Join(templatesDir, extends)
			if err := analyzeTemplate(parentPath); err != nil {
				return err
			}
		}
		return nil
	}

	if err := analyzeTemplate(templatePath); err != nil {
		return nil, err
	}
	return templateAnalysis, nil
}

func parseTemplate(content string) (string, []string, map[string]string, map[string]string) {
	var extends string
	var includes []string
	definedBlocks := make(map[string]string)
	usedBlocks := make(map[string]string)
	extendsRegex := regexp.MustCompile(`{% extends "([^"]+)" %}`)
	if matches := extendsRegex.FindStringSubmatch(content); len(matches) > 1 {
		extends = matches[1]
	}
	includeRegex := regexp.MustCompile(`{% include "([^"]+)" %}`)
	includeMatches := includeRegex.FindAllStringSubmatch(content, -1)
	for _, match := range includeMatches {
		if len(match) > 1 {
			includes = append(includes, match[1])
		}
	}
	blockDefRegex := regexp.MustCompile(`(?s){% block (\w+) %}(.*?){% endblock %}`)
	blockMatches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range blockMatches {
		if len(match) > 2 {
			blockName := match[1]
			definedBlocks[blockName] = match[2]
		}
	}
	blockUsageRegex := regexp.MustCompile(`{{\s*block\("(\w+)"\)\s*}}`)
	usageMatches := blockUsageRegex.FindAllStringSubmatch(content, -1)
	for _, match := range usageMatches {
		if len(match) > 1 {
			blockName := match[1]
			usedBlocks[blockName] = match[0]
		}
	}
	return extends, includes, definedBlocks, usedBlocks
}

func buildTemplate(analysis map[string]TemplateInfo, templatePath string, data map[string]interface{}) (string, error) {
	tmplInfo := analysis[templatePath]
	content := readFileContent(templatePath)
	var buffer bytes.Buffer

	tmpl, err := template.New("base").Parse(content)
	if err != nil {
		return "", err
	}

	for blockName, blockContent := range tmplInfo.DefinedBlocks {
		tmpl = template.Must(tmpl.New(blockName).Parse(blockContent))
	}

	if err := tmpl.Execute(&buffer, data); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func readFileContent(path string) string {
	raw, _ := os.ReadFile(path)
	return string(raw)
}
