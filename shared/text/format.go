package text

// FormatWordList formats a list of words into a human-readable string
// with the given quote character and "or" conjunction.
// Examples:
//
//	FormatWordList([]string{"closed"}, "'") => "'closed'"
//	FormatWordList([]string{"closed", "abandoned"}, "'") => "'closed' or 'abandoned'"
//	FormatWordList([]string{"a", "b", "c"}, `"`) => `"a", "b", or "c"`
func FormatWordList(words []string, quote string) string {
	if len(words) == 0 {
		return ""
	}

	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = quote + w + quote
	}

	if len(quoted) == 1 {
		return quoted[0]
	}
	if len(quoted) == 2 {
		return quoted[0] + " or " + quoted[1]
	}

	result := ""
	for i, q := range quoted {
		if i == len(quoted)-1 {
			result += " or " + q
		} else if i > 0 {
			result += ", " + q
		} else {
			result = q
		}
	}
	return result
}
