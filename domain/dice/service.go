package dice

import (
	"fmt"
	"math/rand"
	"strings"

	rollapi "github.com/justinian/dice"
)

// splitTokens splits a line by commas, but treats commas inside {} as part of
// the same token so that list notation like {Frank; Bill; Joe} is not broken up.
func splitTokens(line string) []string {
	var tokens []string
	depth := 0
	start := 0
	for i, ch := range line {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				tokens = append(tokens, strings.TrimSpace(line[start:i]))
				start = i + 1
			}
		}
	}
	tokens = append(tokens, strings.TrimSpace(line[start:]))
	return tokens
}

// parseList detects a {item; item (2); ...} token and returns a weighted RollResult.
// Returns nil if the token is not a list.
func parseList(token string) *RollResult {
	token = strings.TrimSpace(token)
	if !strings.HasPrefix(token, "{") || !strings.HasSuffix(token, "}") {
		return nil
	}

	inner := strings.TrimSpace(token[1 : len(token)-1])
	pool := buildPool(strings.Split(inner, ";"))

	if len(pool) == 0 {
		return &RollResult{Notation: token, Err: fmt.Errorf("empty list")}
	}

	return &RollResult{
		Notation: token,
		Picked:   pool[rand.Intn(len(pool))],
	}
}

// Roll parses the input string and rolls all dice expressions found within it.
//
// Each non-empty line is treated as a roll group. Lines may optionally start
// with a label followed by ": ". Multiple comma-separated expressions on one
// line are rolled as separate results within the same group.
//
// An optional OracleLookup can be provided to resolve user-defined @oracle
// references. User-defined oracles take precedence over builtins.
//
// Examples:
//
//	"2d6"                         → one unlabeled group, one roll
//	"2d6, 1d8"                    → one unlabeled group, two rolls
//	"Attack: 1d20+5"              → one labeled group, one roll
//	"Attack: 1d20+5, 1d6"         → one labeled group, two rolls
func Roll(input string, oracles ...OracleLookup) []RollGroup {
	var lookup OracleLookup
	if len(oracles) > 0 {
		lookup = oracles[0]
	}
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

		for _, notation := range splitTokens(line) {
			notation = strings.TrimSpace(notation)
			if notation == "" {
				continue
			}

			if listResult := parseList(notation); listResult != nil {
				group.Results = append(group.Results, *listResult)
				continue
			}

			if oracleResult := parseOracle(notation, lookup); oracleResult != nil {
				group.Results = append(group.Results, *oracleResult)
				continue
			}

			notation = strings.ToLower(notation)
			result, _, err := rollapi.Roll(notation)
			if err != nil {
				group.Results = append(group.Results, RollResult{
					Notation: notation,
					Err:      err,
				})
				continue
			}

			rr := RollResult{
				Notation: notation,
				Total:    result.Int(),
			}
			switch r := result.(type) {
			case rollapi.StdResult:
				rr.Rolls = r.Rolls
				rr.Dropped = r.Dropped
			case rollapi.VsResult:
				rr.Rolls = r.Rolls
			case rollapi.FudgeResult:
				rr.Rolls = r.Rolls
			}
			group.Results = append(group.Results, rr)
		}

		if len(group.Results) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}
