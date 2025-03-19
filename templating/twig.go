package templating

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Template struct {
	RawContent string
	Blocks     map[string]string
}

var twigTemplatesDir = "./templates"

func TwigRender(templatePath string, data map[string]interface{}) (string, error) {
	fullTemplatePath := filepath.Join(twigTemplatesDir, templatePath)
	templateAnalysis, err := analyzeTwigTemplates(fullTemplatePath)
	if err != nil {
		return "", err
	}

	builtTemplate, err := buildTwigTemplate(templateAnalysis, fullTemplatePath, data)
	if err != nil {
		return "", err
	}

	finalTemplate, err := cleanAndCheck(builtTemplate)
	if err != nil {
		return "", err
	}

	return finalTemplate, nil
}

type TwigTemplateInfo struct {
	Path          string
	Extends       string
	Includes      []string
	DefinedBlocks map[string]string
	UsedBlocks    map[string]string
	EntryTemplate string
}

func analyzeTwigTemplates(templatePath string) (map[string]TwigTemplateInfo, error) {
	templateAnalysis := make(map[string]TwigTemplateInfo)
	var analyzeTemplate func(path string, isEntry bool) error
	analyzeTemplate = func(path string, isEntry bool) error {
		fmt.Printf("Analyzing template at path: %s\n", path)
		if _, exists := templateAnalysis[path]; exists {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}
		content := string(raw)
		extends, includes, definedBlocks, usedBlocks := parseTwigTemplate(content)
		TwigtemplateInfo := TwigTemplateInfo{
			Path:          path,
			Extends:       extends,
			Includes:      includes,
			DefinedBlocks: definedBlocks,
			UsedBlocks:    usedBlocks,
		}
		if isEntry {
			TwigtemplateInfo.EntryTemplate = path
		}
		templateAnalysis[path] = TwigtemplateInfo
		for _, include := range includes {
			includePath := filepath.Join(twigTemplatesDir, include)
			err := analyzeTemplate(includePath, false)
			if err != nil {
				return err
			}
		}

		if extends != "" {
			parentPath := filepath.Join(twigTemplatesDir, extends)
			err := analyzeTemplate(parentPath, false)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err := analyzeTemplate(templatePath, true)
	if err != nil {
		return nil, err
	}
	return templateAnalysis, nil
}

func parseTwigTemplate(content string) (string, []string, map[string]string, map[string]string) {
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
			definedBlocks[blockName] = strings.TrimSpace(match[2])
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

func buildTwigTemplate(analysis map[string]TwigTemplateInfo, templatePath string, data map[string]interface{}) (string, error) {
	var finalContent strings.Builder
	var renderTemplate func(path string, parentBlocks map[string]string) (string, map[string]string, error)
	renderTemplate = func(path string, parentBlocks map[string]string) (string, map[string]string, error) {
		TwigtemplateInfo, exists := analysis[path]
		if !exists {
			return "", nil, fmt.Errorf("template not found: %s", path)
		}
		content := replaceIncludes(TwigtemplateInfo.Path, TwigtemplateInfo)
		for blockName, usage := range TwigtemplateInfo.UsedBlocks {
			if definedContent, exists := TwigtemplateInfo.DefinedBlocks[blockName]; exists {
				content = strings.ReplaceAll(content, usage, definedContent)
			}
		}
		if TwigtemplateInfo.Extends != "" {
			parentContent, parentBlocks, err := renderTemplate(filepath.Join(twigTemplatesDir, TwigtemplateInfo.Extends), nil)
			if err != nil {
				return "", nil, err
			}
			mergedContent := mergeBlocks(parentContent, parentBlocks, TwigtemplateInfo.DefinedBlocks)
			return mergedContent, TwigtemplateInfo.DefinedBlocks, nil
		} else {
			completeContent := mergeBlocks(content, parentBlocks, TwigtemplateInfo.DefinedBlocks)
			return completeContent, TwigtemplateInfo.DefinedBlocks, nil
		}
	}
	content, _, err := renderTemplate(templatePath, nil)
	if err != nil {
		return "", err
	}
	content = handleVars(content, data)
	entryTwigTemplateInfo := analysis[templatePath]
	if entryTwigTemplateInfo.EntryTemplate != "" {
		for blockName, definedContent := range entryTwigTemplateInfo.DefinedBlocks {
			placeholder := fmt.Sprintf("{%% block %s %%}{%% endblock %%}", blockName)
			content = strings.ReplaceAll(content, placeholder, definedContent)
		}
	}

	finalContent.WriteString(content)
	return finalContent.String(), nil
}

func replaceIncludes(templatePath string, tInfo TwigTemplateInfo) string {
	content := readTwigFileContent(templatePath)

	for _, include := range tInfo.Includes {
		includePath := filepath.Join(filepath.Dir(templatePath), include)
		includedContent := readTwigFileContent(includePath)
		includeTag := `{%` + ` include ` + `"` + include + `"` + ` %}`
		content = strings.Replace(content, includeTag, includedContent, 1)
	}
	return content
}

func readTwigFileContent(path string) string {
	raw, _ := os.ReadFile(path)
	return string(raw)
}

func mergeBlocks(content string, parentBlocks, childBlocks map[string]string) string {
	blockDefRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	matches := blockDefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		blockType := match[2]
		if blockType != "UsedBlock" {
			continue
		}
		blockName := match[1]
		blockPlaceholder := match[0]
		if childContent, exists := childBlocks[blockName]; exists {
			content = strings.Replace(content, blockPlaceholder, childContent, 1)
		} else if parentContent, exists := parentBlocks[blockName]; exists {
			content = strings.Replace(content, blockPlaceholder, parentContent, 1)
		}
	}
	return content
}

func handleVars(content string, data map[string]interface{}) string {
	for key, value := range data {
		placeholder := "{{ " + key + " }}"
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}
	return content
}

func cleanAndCheck(content string) (string, error) {
	unresolvedBlockRegex := regexp.MustCompile(`{% block (\w+) %}(.*?)\{% endblock %}`)
	unresolvedBlocks := unresolvedBlockRegex.FindAllStringSubmatch(content, -1)
	if len(unresolvedBlocks) > 0 {
		for _, block := range unresolvedBlocks {
			content = strings.Replace(content, block[0], "", -1)
		}
	}
	return content, nil
}
