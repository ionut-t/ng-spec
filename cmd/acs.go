package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

func parseAcs(acsText string) string {
	lines := strings.Split(acsText, "\n")
	var result strings.Builder

	level1Regex := regexp.MustCompile(`^\s*(\d+)\.\s+(.+)$`)
	level2Regex := regexp.MustCompile(`^\s*([a-z])\.\s+(.+)$`)
	level3Regex := regexp.MustCompile(`^\s*([ivx]+)\.\s+(.+)$`)

	// Keep track of the current context
	var currentLevel1, currentLevel2 string
	indentLevel := 0

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		if matches := level1Regex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			// Close previous blocks if any
			if currentLevel2 != "" {
				result.WriteString(getIndentation(1) + "});\n\n")
				currentLevel2 = ""
			}
			if currentLevel1 != "" {
				result.WriteString("});\n\n")
			}

			// Start new level 1 block
			title := sanitizeTitle(matches[2])
			result.WriteString(fmt.Sprintf("describe('%s', () => {\n", title))
			currentLevel1 = title
			indentLevel = 1

		} else if matches := level2Regex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			// Close previous level 2 block if any
			if currentLevel2 != "" {
				result.WriteString(getIndentation(1) + "});\n\n")
			}

			title := sanitizeTitle(matches[2])

			// Check if this is a test case or a nested describe
			if strings.Contains(title, "describe") {
				title = strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(title, "describe")), ")")
				title = strings.Trim(title, "()\"' ")
				result.WriteString(fmt.Sprintf("%sdescribe('%s', () => {\n", getIndentation(indentLevel), title))
				currentLevel2 = title
				indentLevel = 2
			} else {
				result.WriteString(fmt.Sprintf("%sit('should %s', async () => {\n",
					getIndentation(indentLevel),
					lcFirst(title)))
				result.WriteString(fmt.Sprintf("%s// TODO: Implement test\n", getIndentation(indentLevel+1)))
				result.WriteString(fmt.Sprintf("%s});\n\n", getIndentation(indentLevel)))
			}

		} else if matches := level3Regex.FindStringSubmatch(trimmedLine); len(matches) > 0 {
			title := sanitizeTitle(matches[2])
			result.WriteString(fmt.Sprintf("%sit('should %s', async () => {\n",
				getIndentation(indentLevel),
				lcFirst(title)))
			result.WriteString(fmt.Sprintf("%s// TODO: Implement test\n", getIndentation(indentLevel+1)))
			result.WriteString(fmt.Sprintf("%s});\n\n", getIndentation(indentLevel)))
		}
	}

	// Close any open blocks
	if currentLevel2 != "" {
		result.WriteString(getIndentation(1) + "});\n")
	}

	if currentLevel1 != "" {
		result.WriteString("});\n")
	}

	return result.String()
}

func getIndentation(level int) string {
	return strings.Repeat("\t", level)
}

func sanitizeTitle(title string) string {
	if strings.Contains(strings.ToLower(title), "describe (") {
		return title
	}

	title = strings.TrimSpace(title)
	return title
}

func lcFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func integrateAcsWithTemplate(templateContent, acsLink, acsBlocks string) string {
	endBlock := "});"

	content := strings.TrimSuffix(strings.TrimSpace(templateContent), endBlock)

	if acsLink != "" {
		content = strings.Replace(content, "TODO: Link ACs tickets here", acsLink, 1)
	}

	content += "\n" + acsBlocks
	content += endBlock

	return content
}
