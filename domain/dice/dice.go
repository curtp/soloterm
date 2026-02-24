package dice

type RollGroup struct {
	Label   string // empty if unlabeled
	Results []RollResult
}

type RollResult struct {
	Notation string
	Total    int
	Rolls    []int // kept dice values (or all dice if no keep/drop)
	Dropped  []int // dropped dice values (nil if no keep/drop)
	Err      error // per-roll error, not fatal to the group
}
