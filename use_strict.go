package assetmin

import "bytes"

// stripLeadingUseStrict removes a leading "use strict" directive (with or
// without trailing semicolon and optional surrounding whitespace/newline)
// from the provided JS file content. It only removes a directive at the very
// start of the file. The function returns a new byte slice.
func stripLeadingUseStrict(b []byte) []byte {
	if len(b) == 0 {
		return b
	}

	// Consider possible variants: "use strict"; or 'use strict';
	// Trim left whitespace and newlines first to detect directive at beginning.
	trimmed := bytes.TrimLeft(b, "\t \n\r\f\v")
	lower := bytes.ToLower(trimmed)

	// Check double-quoted
	if bytes.HasPrefix(lower, []byte("\"use strict\"")) || bytes.HasPrefix(lower, []byte("'use strict'")) {
		// Find end of the directive
		// advance past the closing quote
		i := 0
		for i < len(trimmed) && trimmed[i] != '\n' {
			// stop at first newline; we'll strip up to there
			i++
		}
		// If the first line didn't contain a newline we still want to remove the
		// directive plus any following semicolon and whitespace. So compute a
		// conservative end index by scanning for the first character after the
		// closing quote and optional semicolon.
		firstLine := trimmed[:i]
		// find the index after the closing quote
		end := -1
		for j := 0; j < len(firstLine); j++ {
			ch := firstLine[j]
			if ch == '"' || ch == '\'' {
				// attempt to find matching closing quote after this
				// naive but good enough: find next same quote
				for k := j + 1; k < len(firstLine); k++ {
					if firstLine[k] == ch {
						// position after closing quote
						end = k + 1
						break
					}
				}
				break
			}
		}
		if end == -1 {
			// malformed, return original
			return b
		}
		// skip trailing semicolon and whitespace/newline
		for end < len(firstLine) && (firstLine[end] == ';' || firstLine[end] == ' ' || firstLine[end] == '\t' || firstLine[end] == '\r') {
			end++
		}
		// compute remainder starting from trimmed[end:]
		remainder := trimmed[end:]
		// Now reattach any leading whitespace that was trimmed originally
		// that existed before the directive. We previously removed all left
		// whitespace, so just return remainder (no leading whitespace)
		return remainder
	}

	return b
}
