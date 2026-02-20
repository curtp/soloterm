package dice

type RollGroup struct {
	Label   string // empty if unlabeled
	Results []RollResult
}

type RollResult struct {
	Notation  string
	Total     int
	Breakdown string
	Err       error // per-roll error, not fatal to the group
}
