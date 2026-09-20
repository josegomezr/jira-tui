// from: https://github.com/mgirouard/md/blob/ae53d1f61fa1cb11e493a35c7d5145b53630c9c7/cmd/jira2md/main.go
// Copyright 2026 Mike Girouard

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.

// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.

// 3. Neither the name of the copyright holder nor the names of its contributors
//    may be used to endorse or promote products derived from this software
//    without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS “AS IS” AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Converter converts Jira Wiki Markup to Markdown
type Converter struct {
	input  *bufio.Scanner
	output *bytes.Buffer
}

// NewConverter creates a new Jira to Markdown converter
func NewConverter(r io.Reader) *Converter {
	return &Converter{
		input:  bufio.NewScanner(r),
		output: &bytes.Buffer{},
	}
}

// Convert performs the conversion
func (c *Converter) Convert() (string, error) {
	var lines []string
	for c.input.Scan() {
		lines = append(lines, c.input.Text())
	}
	if err := c.input.Err(); err != nil {
		return "", err
	}

	// Process lines
	i := 0
	for i < len(lines) {
		line := lines[i]

		// Handle empty lines
		if line == "" {
			c.output.WriteString("\n")
			i++
			continue
		}

		// Handle code blocks
		if strings.HasPrefix(line, "{code") {
			i = c.processCodeBlock(lines, i)
			continue
		}

		// Handle horizontal rule
		if line == "----" {
			c.output.WriteString("---\n")
			i++
			continue
		}

		// Handle headings
		if matched, _ := regexp.MatchString(`^h[1-6]\.`, line); matched {
			c.processHeading(line)
			i++
			continue
		}

		// Handle tables
		if strings.HasPrefix(line, "||") {
			i = c.processTable(lines, i)
			continue
		}

		// Handle blockquote
		if strings.HasPrefix(line, "{quote}") {
			i = c.processBlockquote(lines, i)
			continue
		}

		// Handle lists
		if c.isListItem(line) {
			i = c.processList(lines, i)
			continue
		}

		// Handle panels, info, tip, warning, note boxes
		if strings.HasPrefix(line, "{panel}") || strings.HasPrefix(line, "{info}") ||
			strings.HasPrefix(line, "{tip}") || strings.HasPrefix(line, "{warning}") ||
			strings.HasPrefix(line, "{note}") {
			i = c.processPanel(lines, i)
			continue
		}

		// Regular paragraph
		c.processParagraph(line)
		i++
	}

	return c.output.String(), nil
}

// processHeading converts Jira headings to markdown format
func (c *Converter) processHeading(line string) {
	// Extract heading level
	re := regexp.MustCompile(`^h([1-6])\.\s*(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		level := matches[1]
		text := c.processInline(matches[2])
		c.output.WriteString(strings.Repeat("#", mustAtoi(level)) + " " + text + "\n")
	}
}

// mustAtoi converts string to int, returns 0 on error
func mustAtoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// isListItem checks if a line is a Jira list item
func (c *Converter) isListItem(line string) bool {
	if len(line) == 0 {
		return false
	}
	// Check for * or # at the start
	return regexp.MustCompile(`^([*#]+)\s`).MatchString(line)
}

// getListInfo returns the list marker and content
func (c *Converter) getListInfo(line string) (marker string, content string) {
	re := regexp.MustCompile(`^([*#]+)\s*(.*)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", line
}

// processList converts Jira lists to markdown format
func (c *Converter) processList(lines []string, start int) int {
	i := start
	for i < len(lines) {
		line := lines[i]

		// Empty line ends the list
		if line == "" {
			break
		}

		// Check if this is still part of a list
		marker, content := c.getListInfo(line)
		if marker == "" {
			break
		}

		level := len(marker)
		isOrdered := marker[0] == '#'

		indent := strings.Repeat("  ", level-1)
		if isOrdered {
			c.output.WriteString(indent + "1. " + c.processInline(content) + "\n")
		} else {
			c.output.WriteString(indent + "- " + c.processInline(content) + "\n")
		}

		i++
	}
	return i
}

// processBlockquote converts Jira blockquote to markdown format
func (c *Converter) processBlockquote(lines []string, start int) int {
	line := lines[start]

	// Content may start on the same line as the opening {quote}
	var content []string
	openingContent := strings.TrimPrefix(line, "{quote}")
	if openingContent != "" {
		if strings.HasSuffix(openingContent, "{quote}") {
			openingContent = strings.TrimSuffix(openingContent, "{quote}")
		}
		content = append(content, openingContent)
	}

	i := start + 1
	for i < len(lines) {
		if lines[i] == "{quote}" {
			i++
			break
		}
		// Closing {quote} may appear at end of a content line
		if strings.HasSuffix(lines[i], "{quote}") {
			content = append(content, strings.TrimSuffix(lines[i], "{quote}"))
			i++
			break
		}
		content = append(content, lines[i])
		i++
	}

	for _, contentLine := range content {
		c.output.WriteString("> " + c.processInline(contentLine) + "\n")
	}
	return i
}

// processCodeBlock handles Jira code blocks
func (c *Converter) processCodeBlock(lines []string, start int) int {
	line := lines[start]

	// Extract language if specified
	lang := ""
	re := regexp.MustCompile(`{code:([^}]+)}`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		lang = matches[1]
	}

	i := start + 1
	var codeLines []string
	for i < len(lines) {
		if lines[i] == "{code}" {
			i++
			break
		}
		codeLines = append(codeLines, lines[i])
		i++
	}

	if lang != "" {
		c.output.WriteString("```" + lang + "\n")
	} else {
		c.output.WriteString("```\n")
	}
	c.output.WriteString(strings.Join(codeLines, "\n"))
	c.output.WriteString("\n```\n")

	return i
}

// processTable converts Jira tables to markdown format
func (c *Converter) processTable(lines []string, start int) int {
	i := start

	// Parse header (||col1||col2||)
	headerLine := lines[i]
	headers := c.parseJiraTableRow(headerLine, true)
	i++

	// Collect rows
	var rows [][]string
	for i < len(lines) && strings.HasPrefix(lines[i], "|") && !strings.HasPrefix(lines[i], "||") {
		rows = append(rows, c.parseJiraTableRow(lines[i], false))
		i++
	}

	// Output markdown table
	// Header
	c.output.WriteString("| " + strings.Join(headers, " | ") + " |\n")

	// Separator
	sep := make([]string, len(headers))
	for j := range sep {
		sep[j] = "---"
	}
	c.output.WriteString("| " + strings.Join(sep, " | ") + " |\n")

	// Rows
	for _, row := range rows {
		// Pad row if needed
		for len(row) < len(headers) {
			row = append(row, "")
		}
		c.output.WriteString("| " + strings.Join(row, " | ") + " |\n")
	}

	return i
}

// parseJiraTableRow parses a Jira table row
func (c *Converter) parseJiraTableRow(line string, isHeader bool) []string {
	var cells []string

	// For Jira, cells are separated by |
	// Header uses || as delimiters, data rows use |
	trimmed := strings.TrimSpace(line)

	if isHeader {
		// Remove leading ||
		trimmed = strings.TrimPrefix(trimmed, "||")
		// Split by ||
		parts := strings.Split(trimmed, "||")
		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				cells = append(cells, c.processInline(part))
			}
		}
	} else {
		// Remove leading |
		trimmed = strings.TrimPrefix(trimmed, "|")
		// Split by |
		parts := strings.Split(trimmed, "|")
		for _, part := range parts {
			if part = strings.TrimSpace(part); part != "" {
				cells = append(cells, c.processInline(part))
			}
		}
	}

	return cells
}

// processPanel converts Jira panels to markdown blockquotes or notes
func (c *Converter) processPanel(lines []string, start int) int {
	line := lines[start]

	// Extract panel type
	panelType := "panel"
	re := regexp.MustCompile(`{([^}]+)}`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		panelType = matches[1]
	}

	i := start + 1
	var content []string
	for i < len(lines) {
		if lines[i] == "{"+panelType+"}" {
			i++
			break
		}
		content = append(content, lines[i])
		i++
	}

	// Convert to markdown blockquote with panel type
	for _, line := range content {
		if panelType == "panel" {
			c.output.WriteString("> " + c.processInline(line) + "\n")
		} else {
			c.output.WriteString("> " + panelType + ": " + c.processInline(line) + "\n")
		}
	}

	return i
}

// processParagraph handles regular paragraphs
func (c *Converter) processParagraph(line string) {
	line = c.processInline(line)
	c.output.WriteString(line + "\n")
}

// processInline processes inline formatting
func (c *Converter) processInline(text string) string {
	var result strings.Builder
	i := 0
	for i < len(text) {
		// Skipped: Strikethrough first: -text- -> ~~text~~
		if false && text[i] == '-' {
			end := strings.Index(text[i+1:], "-")
			if end != -1 && end > 0 {
				result.WriteString("~~")
				result.WriteString(text[i+1 : i+1+end])
				result.WriteString("~~")
				i += end + 2
				continue
			}
		}

		// Images: !url|alt=text! or !url!
		if text[i] == '!' {
			end := strings.Index(text[i+1:], "!")
			if end != -1 {
				content := text[i+1 : i+1+end]
				if pipeIdx := strings.Index(content, "|alt="); pipeIdx != -1 {
					url := content[:pipeIdx]
					alt := content[pipeIdx+5:]
					result.WriteString(fmt.Sprintf("![%s](%s)", alt, url))
				} else {
					result.WriteString(fmt.Sprintf("![](%s)", content))
				}
				i += end + 2
				continue
			}
		}

		// Links: [text|url] or [url]
		if text[i] == '[' {
			end := strings.Index(text[i+1:], "]")
			if end != -1 {
				content := text[i+1 : i+1+end]
				if pipeIdx := strings.Index(content, "|"); pipeIdx != -1 {
					linkText := content[:pipeIdx]
					url := content[pipeIdx+1:]
					result.WriteString(fmt.Sprintf("[%s](%s)", linkText, url))
				} else {
					if content[0] == '~' {
						result.WriteString(fmt.Sprintf("@[%s]", content))
					}else{
						result.WriteString(fmt.Sprintf("%s", content))
					}
				}
				i += end + 2
				continue
			}
		}

		// Code: {{code}}
		if i < len(text)-1 && text[i] == '{' && text[i+1] == '{' {
			end := strings.Index(text[i+2:], "}}")
			if end != -1 {
				result.WriteString("`")
				result.WriteString(text[i+2 : i+2+end])
				result.WriteString("`")
				i += end + 4
				continue
			}
		}

		// Bold+Italic: *_text_* -> ***text***
		// This is a Jira pattern for bold+italic
		if i < len(text)-1 && text[i] == '*' && text[i+1] == '_' {
			end := strings.Index(text[i+2:], "_*")
			if end != -1 {
				result.WriteString("***")
				result.WriteString(text[i+2 : i+2+end])
				result.WriteString("***")
				i += end + 4
				continue
			}
		}

		// Bold: *text* -> **text**
		if text[i] == '*' {
			end := strings.Index(text[i+1:], "*")
			if end != -1 {
				result.WriteString("**")
				result.WriteString(text[i+1 : i+1+end])
				result.WriteString("**")
				i += end + 2
				continue
			}
		}

		// Italic: _text_ -> *text*
		if text[i] == '_' {
			end := strings.Index(text[i+1:], "_")
			if end != -1 {
				result.WriteString("*")
				result.WriteString(text[i+1 : i+1+end])
				result.WriteString("*")
				i += end + 2
				continue
			}
		}

		result.WriteByte(text[i])
		i++
	}
	return result.String()
}
