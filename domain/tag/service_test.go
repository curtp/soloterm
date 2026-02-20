package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// extractRecentTags doesn't use sessionRepo, so a bare &Service{} is sufficient.
func newTestService() *Service {
	return &Service{}
}

func TestExtractRecentTags_DiceBreakdownsIgnored(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name     string
		contents []string
	}{
		{
			name:     "single number in brackets",
			contents: []string{"Rolled 1d6: [4]"},
		},
		{
			name:     "multiple numbers in brackets",
			contents: []string{"Rolled 3d6: [3 3 3]"},
		},
		{
			name:     "dice breakdown with kept and dropped",
			contents: []string{"Rolled 4d6kh3: 9 [3 4 6] ([1])"},
		},
		{
			name:     "multi-roll result inserted into log",
			contents: []string{"Attack: 1d20+5 = 18 [13]\nDamage: 2d6 = 7 [3 4]"},
		},
		{
			name:     "numbers with extra whitespace",
			contents: []string{"Result: [ 12 ]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := svc.extractRecentTags(tc.contents, nil)
			assert.Empty(t, tags, "dice breakdown values should not produce tags")
		})
	}
}

func TestExtractRecentTags_RealTagsStillExtracted(t *testing.T) {
	svc := newTestService()

	tests := []struct {
		name          string
		contents      []string
		wantLabels    []string
		wantTemplates []string
	}{
		{
			name:          "simple labeled tag",
			contents:      []string{"Entered [L:Tavern | cozy]"},
			wantLabels:    []string{"L:Tavern"},
			wantTemplates: []string{"[L:Tavern | cozy]"},
		},
		{
			name:          "tag without data section",
			contents:      []string{"Met [N:Aldric]"},
			wantLabels:    []string{"N:Aldric"},
			wantTemplates: []string{"[N:Aldric]"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := svc.extractRecentTags(tc.contents, nil)
			assert.Len(t, tags, len(tc.wantLabels))
			for i, tag := range tags {
				assert.Equal(t, tc.wantLabels[i], tag.Label)
				assert.Equal(t, tc.wantTemplates[i], tag.Template)
			}
		})
	}
}

func TestExtractRecentTags_MixedContentWithDiceResults(t *testing.T) {
	svc := newTestService()

	// Simulates a session log where dice results were inserted alongside real tags
	content := `Entered the dungeon [L:Dungeon | dark].
Met [N:Skeleton | HP: 3].
Attacked: 1d20+3 = 17 [14]
Damage: 2d6 = 9 [4 5]`

	tags := svc.extractRecentTags([]string{content}, nil)

	labels := make([]string, len(tags))
	for i, t := range tags {
		labels[i] = t.Label
	}

	assert.Len(t, tags, 2)
	assert.Contains(t, labels, "L:Dungeon")
	assert.Contains(t, labels, "N:Skeleton")
}
