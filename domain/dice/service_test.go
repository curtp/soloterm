package dice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoll_Structure(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantGroups  int
		wantResults []int // number of results per group
		wantLabels  []string
	}{
		{
			name:        "empty input",
			input:       "",
			wantGroups:  0,
			wantResults: nil,
			wantLabels:  nil,
		},
		{
			name:        "blank lines only",
			input:       "\n\n   \n",
			wantGroups:  0,
			wantResults: nil,
			wantLabels:  nil,
		},
		{
			name:        "single unlabeled roll",
			input:       "2d6",
			wantGroups:  1,
			wantResults: []int{1},
			wantLabels:  []string{""},
		},
		{
			name:        "single labeled roll",
			input:       "Attack: 1d20",
			wantGroups:  1,
			wantResults: []int{1},
			wantLabels:  []string{"Attack"},
		},
		{
			name:        "multiple unlabeled rolls on one line",
			input:       "2d6, 1d8",
			wantGroups:  1,
			wantResults: []int{2},
			wantLabels:  []string{""},
		},
		{
			name:        "labeled multiple rolls on one line",
			input:       "Attack: 1d20+5, 1d6",
			wantGroups:  1,
			wantResults: []int{2},
			wantLabels:  []string{"Attack"},
		},
		{
			name:        "multiple lines become multiple groups",
			input:       "Attack: 1d20\nDamage: 2d6",
			wantGroups:  2,
			wantResults: []int{1, 1},
			wantLabels:  []string{"Attack", "Damage"},
		},
		{
			name:        "mixed labeled and unlabeled lines",
			input:       "1d6\nAttack: 1d20",
			wantGroups:  2,
			wantResults: []int{1, 1},
			wantLabels:  []string{"", "Attack"},
		},
		{
			name:        "blank lines between rolls are skipped",
			input:       "1d6\n\n2d8\n",
			wantGroups:  2,
			wantResults: []int{1, 1},
			wantLabels:  []string{"", ""},
		},
		{
			name:        "whitespace trimmed from label",
			input:       "  Attack  :  1d20  ",
			wantGroups:  1,
			wantResults: []int{1},
			wantLabels:  []string{"Attack"},
		},
		{
			name:        "three rolls on one labeled line",
			input:       "Pool: 3d6, 1d8, 1d4",
			wantGroups:  1,
			wantResults: []int{3},
			wantLabels:  []string{"Pool"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			groups := Roll(tc.input)

			require.Len(t, groups, tc.wantGroups)

			for i, group := range groups {
				assert.Equal(t, tc.wantLabels[i], group.Label, "group %d label", i)
				assert.Len(t, group.Results, tc.wantResults[i], "group %d result count", i)
			}
		})
	}
}

func TestRoll_ResultValues(t *testing.T) {
	t.Run("1d6 total is in range 1-6", func(t *testing.T) {
		for range 20 {
			groups := Roll("1d6")
			require.Len(t, groups, 1)
			require.Len(t, groups[0].Results, 1)
			result := groups[0].Results[0]
			assert.NoError(t, result.Err)
			assert.GreaterOrEqual(t, result.Total, 1)
			assert.LessOrEqual(t, result.Total, 6)
		}
	})

	t.Run("2d6 total is in range 2-12", func(t *testing.T) {
		for range 20 {
			groups := Roll("2d6")
			result := groups[0].Results[0]
			assert.NoError(t, result.Err)
			assert.GreaterOrEqual(t, result.Total, 2)
			assert.LessOrEqual(t, result.Total, 12)
		}
	})

	t.Run("1d20+5 total is in range 6-25", func(t *testing.T) {
		for range 20 {
			groups := Roll("1d20+5")
			result := groups[0].Results[0]
			assert.NoError(t, result.Err)
			assert.GreaterOrEqual(t, result.Total, 6)
			assert.LessOrEqual(t, result.Total, 25)
		}
	})

	t.Run("notation is stored on result", func(t *testing.T) {
		groups := Roll("Attack: 2d6+3")
		result := groups[0].Results[0]
		assert.Equal(t, "2d6+3", result.Notation)
	})

	t.Run("rolls is non-empty on success", func(t *testing.T) {
		groups := Roll("2d6")
		result := groups[0].Results[0]
		assert.NoError(t, result.Err)
		assert.NotEmpty(t, result.Rolls)
	})
}

func TestRoll_InvalidNotation(t *testing.T) {
	t.Run("invalid notation sets Err on result", func(t *testing.T) {
		groups := Roll("notdice")
		require.Len(t, groups, 1)
		require.Len(t, groups[0].Results, 1)
		result := groups[0].Results[0]
		assert.Error(t, result.Err)
		assert.Zero(t, result.Total)
		assert.Empty(t, result.Rolls)
	})

	t.Run("invalid notation does not prevent other rolls in same group", func(t *testing.T) {
		groups := Roll("1d6, notdice, 1d8")
		require.Len(t, groups, 1)
		require.Len(t, groups[0].Results, 3)
		assert.NoError(t, groups[0].Results[0].Err)
		assert.Error(t, groups[0].Results[1].Err)
		assert.NoError(t, groups[0].Results[2].Err)
	})

	t.Run("invalid notation on one line does not prevent other lines", func(t *testing.T) {
		groups := Roll("notdice\n1d6")
		require.Len(t, groups, 2)
		assert.Error(t, groups[0].Results[0].Err)
		assert.NoError(t, groups[1].Results[0].Err)
	})
}
