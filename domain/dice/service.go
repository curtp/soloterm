package dice

import (
	"strings"

	rollapi "github.com/justinian/dice"
)

// Roll parses the input string and rolls all dice expressions found within it.
//
// Each non-empty line is treated as a roll group. Lines may optionally start
// with a label followed by ": ". Multiple comma-separated expressions on one
// line are rolled as separate results within the same group.
//
// Examples:
//
//	"2d6"                         → one unlabeled group, one roll
//	"2d6, 1d8"                    → one unlabeled group, two rolls
//	"Attack: 1d20+5"              → one labeled group, one roll
//	"Attack: 1d20+5, 1d6"         → one labeled group, two rolls
func Roll(input string) []RollGroup {
	groups := make([]RollGroup, 0)

	for line := range strings.SplitSeq(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		group := RollGroup{}

		// Optional "Label: " prefix
		if idx := strings.Index(line, ": "); idx != -1 {
			group.Label = strings.TrimSpace(line[:idx])
			line = strings.TrimSpace(line[idx+2:])
		}

		for notation := range strings.SplitSeq(line, ",") {
			notation = strings.TrimSpace(notation)
			if notation == "" {
				continue
			}

			result, _, err := rollapi.Roll(notation)
			if err != nil {
				group.Results = append(group.Results, RollResult{
					Notation: notation,
					Err:      err,
				})
				continue
			}

			group.Results = append(group.Results, RollResult{
				Notation:  notation,
				Total:     result.Int(),
				Breakdown: result.String(),
			})
		}

		if len(group.Results) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}
