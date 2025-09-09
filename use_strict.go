package assetmin

import "strings"

// stripLeadingUseStrict removes the first "use strict" directive found at the
// beginning of a JavaScript file, even if preceded by comments or whitespace.
// It removes the directive along with its optional semicolon and following whitespace/newline.
func stripLeadingUseStrict(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	content := string(b)
	patterns := []string{
		"\"use strict\"",
		"'use strict'",
	}

	// Find the first occurrence of any "use strict" pattern
	var foundPos = -1
	var foundPattern string

	for _, pattern := range patterns {
		pos := strings.Index(content, pattern)
		if pos != -1 && (foundPos == -1 || pos < foundPos) {
			foundPos = pos
			foundPattern = pattern
		}
	}

	// If no "use strict" found, return original
	if foundPos == -1 {
		return b
	}

	// Check if this "use strict" is at the logical beginning of the file
	// (i.e., only comments, whitespace, or nothing before it)
	beforeDirective := content[:foundPos]
	if !isOnlyCommentsAndWhitespace(beforeDirective) {
		return b // Not at the beginning, don't remove
	}

	// Found "use strict" at the beginning, remove it
	pos := foundPos + len(foundPattern)

	// Skip optional semicolon
	if pos < len(content) && content[pos] == ';' {
		pos++
	}

	// Skip whitespace after semicolon but stop at newline
	for pos < len(content) && (content[pos] == ' ' || content[pos] == '\t' || content[pos] == '\r') {
		pos++
	}

	// Skip one newline if present
	if pos < len(content) && content[pos] == '\n' {
		pos++
	}

	// Return the remainder
	if pos < len(content) {
		return []byte(content[pos:])
	}
	return []byte{}
}

// isOnlyCommentsAndWhitespace checks if a string contains only comments and whitespace
func isOnlyCommentsAndWhitespace(s string) bool {
	i := 0
	for i < len(s) {
		ch := s[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Check for single-line comment //
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '/' {
			// Skip to end of line
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}

		// Check for multi-line comment /* */
		if i+1 < len(s) && s[i] == '/' && s[i+1] == '*' {
			i += 2
			// Find closing */
			for i+1 < len(s) {
				if s[i] == '*' && s[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}

		// Found non-comment, non-whitespace character
		return false
	}

	return true
}
