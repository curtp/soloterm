package dice

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
)

var listWeightRegex = regexp.MustCompile(`^(.+?)\s*\((\d+)\)$`)

// buildPool expands a slice of entries (which may include weight suffixes like
// "Torso (2)") into a weighted pool ready for random selection.
func buildPool(entries []string) []string {
	var pool []string
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		weight := 1
		if m := listWeightRegex.FindStringSubmatch(entry); m != nil {
			if w, err := strconv.Atoi(m[2]); err == nil && w > 0 {
				weight = w
				entry = strings.TrimSpace(m[1])
			}
		}
		for range weight {
			pool = append(pool, entry)
		}
	}
	return pool
}

// OracleLookup is implemented by any type that can resolve oracle names to entry lists.
type OracleLookup interface {
	Lookup(name string) ([]string, bool)
}

// parseOracle detects an @name token and returns a RollResult picked from the
// matching oracle. User-defined oracles are checked first, then builtins.
// Returns nil if the token is not an @name reference.
func parseOracle(token string, lookup OracleLookup) *RollResult {
	if !strings.HasPrefix(token, "@") {
		return nil
	}
	name := strings.ToLower(strings.TrimSpace(token[1:]))

	if lookup != nil {
		if entries, ok := lookup.Lookup(name); ok {
			pool := buildPool(entries)
			if len(pool) > 0 {
				return &RollResult{Notation: token, Picked: pool[rand.Intn(len(pool))]}
			}
		}
	}

	return &RollResult{Notation: token, Err: fmt.Errorf("unknown oracle: %s", name)}
}
