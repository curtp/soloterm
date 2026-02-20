package tag

import (
	"regexp"
	"soloterm/domain/session"
	"sort"
	"strings"
)

// Service handles tag-related business logic
type Service struct {
	sessionRepo *session.Repository
}

// NewService creates a new tag service
func NewService(sessionRepo *session.Repository) *Service {
	return &Service{sessionRepo: sessionRepo}
}

// LoadTagsForGame loads both configured tags and recent tags from sessions
// Returns configured tags first, then recent tags extracted from session content
func (s *Service) LoadTagsForGame(gameID int64, configTags []TagType, excludeWords []string) ([]TagType, error) {
	// Sort the config tags by label
	sort.Slice(configTags, func(i, j int) bool {
		return configTags[i].Label < configTags[j].Label
	})

	// If no game selected, just return configured tags
	if gameID == 0 {
		return configTags, nil
	}

	// Get all session content for the game
	contents, err := s.sessionRepo.GetAllContentForGame(gameID)
	if err != nil {
		// Return configured tags even if loading fails
		return configTags, nil
	}

	// Extract recent tags from session content
	recentTags := s.extractRecentTags(contents, excludeWords)

	// Combine: configured tags first, then recent tags
	allTags := make([]TagType, 0, len(configTags)+len(recentTags))
	allTags = append(allTags, configTags...)
	allTags = append(allTags, recentTags...)

	return allTags, nil
}

// extractRecentTags extracts tags from session content, deduplicates by type, keeps most recent
func (s *Service) extractRecentTags(contents []string, excludeWords []string) []TagType {
	// Map of tag type (identifier) -> most recent full tag
	// We iterate in reverse order (most recent first) to keep the newest
	tagMap := make(map[string]string)

	// Track identifiers that should be excluded (any identifier where we find an exclude word)
	excludedIdentifiers := make(map[string]bool)

	// Regex to match tags: [TagIdentifier | data]
	// Captures tag identifier and optional data section separately
	// The data section allows nested [...] sequences (e.g. for tracking boxes like [ ])
	tagRegex := regexp.MustCompile(`\[([^\]|]+)(\|[^\[\]]*(?:\[[^\]]*\][^\[\]]*)*)?\]`)

	// Dice breakdown values like [3 3 3] or [1] should not be treated as tags
	numericOnlyRegex := regexp.MustCompile(`^[\d\s]+$`)

	// Process contents in reverse order (newest first)
	for i := len(contents) - 1; i >= 0; i-- {
		content := contents[i]

		// Find all tags in this content
		matches := tagRegex.FindAllStringSubmatch(content, -1)

		// Iterate matches in reverse so the last occurrence in a session wins
		for j := len(matches) - 1; j >= 0; j-- {
			match := matches[j]
			if len(match) >= 2 {
				// match[1] is the tag identifier (everything from [ to first | or ])
				identifier := strings.TrimSpace(match[1])
				if numericOnlyRegex.MatchString(identifier) {
					continue
				}
				// match[0] is the entire tag including brackets
				fullTag := match[0]

				// Check if there's a data section (match[2]) and if it contains any exclude words
				if len(match) >= 3 && match[2] != "" && len(excludeWords) > 0 {
					dataSection := strings.ToLower(match[2])
					for _, word := range excludeWords {
						if strings.Contains(dataSection, strings.ToLower(word)) {
							// Mark this identifier for exclusion
							excludedIdentifiers[identifier] = true
							// Remove from map if we already added it
							delete(tagMap, identifier)
							break
						}
					}
				}

				// Only store if we haven't seen this identifier yet and it's not excluded
				if _, exists := tagMap[identifier]; !exists && !excludedIdentifiers[identifier] {
					tagMap[identifier] = fullTag
				}
			}
		}
	}

	// Convert map to slice of TagTypes
	recentTags := make([]TagType, 0, len(tagMap))
	for prefix, fullTag := range tagMap {
		recentTags = append(recentTags, TagType{
			Label:    prefix,
			Template: fullTag,
		})
	}

	// Sort by prefix (alphabetically)
	sort.Slice(recentTags, func(i, j int) bool {
		return recentTags[i].Label < recentTags[j].Label
	})

	return recentTags
}
