package oracle

import (
	"fmt"
	"strings"
)

// Service handles oracle business logic
type Service struct {
	repo *Repository
}

// NewService creates a new oracle service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Save validates and saves a oracle (create or update)
func (s *Service) Save(g *Oracle) (*Oracle, error) {
	// Validate
	validator := g.Validate()
	if validator.HasErrors() {
		return nil, validator
	}

	// Save to database
	err := s.repo.Save(g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// SaveContent updates the content on a oracle
func (s *Service) SaveContent(oracleID int64, content string) error {
	return s.repo.SaveContent(oracleID, content)
}

// Delete removes a oracle by ID
func (s *Service) Delete(id int64) error {
	_, err := s.repo.Delete(id)
	return err
}

// GetByID retrieves a oracle by ID
func (s *Service) GetByID(id int64) (*Oracle, error) {
	return s.repo.GetByID(id)
}

// GetAll retrieves all oracles
func (s *Service) GetAll() ([]*Oracle, error) {
	return s.repo.GetAll()
}

// GetCategoryInfo returns per-category sort data for use by the oracle form
func (s *Service) GetCategoryInfo() ([]*CategoryInfo, error) {
	return s.repo.GetCategoryInfo()
}

// GetTableHints returns "Category/Name" strings matching the given prefix,
// for display in the dice roller as the user types after "@".
// An empty prefix returns up to 20 tables. Returns nil on error or no match.
func (s *Service) GetTableHints(prefix string) []string {
	oracles, err := s.repo.FindByPrefix(prefix)
	if err != nil || len(oracles) == 0 {
		return nil
	}
	hints := make([]string, len(oracles))
	for i, o := range oracles {
		hints[i] = o.Category + "/" + o.Name
	}
	return hints
}

// Reorder moves an oracle or its entire category up or down.
// direction: -1 = up, +1 = down.
//
//   - If oracleID is 0, moves the category identified by categoryName as a block.
//   - If oracleID > 0, moves that oracle within its category only.
//
// Returns the ID of the moved oracle (or the first oracle in the moved category),
// 0 on a boundary no-op, or an error.
func (s *Service) Reorder(oracleID int64, categoryName string, direction int) (int64, error) {
	oracles, err := s.repo.GetAll()
	if err != nil {
		return 0, err
	}

	if oracleID == 0 {
		return s.reorderCategory(oracles, categoryName, direction)
	}
	return s.reorderOracle(oracles, oracleID, direction)
}

func (s *Service) reorderCategory(oracles []*Oracle, categoryName string, direction int) (int64, error) {
	// Find the category_position of the named category and an adjacent category to swap with
	currPos := -1
	var firstID int64
	for _, o := range oracles {
		if o.Category == categoryName {
			if currPos == -1 {
				currPos = o.CategoryPosition
				firstID = o.ID
			}
			break
		}
	}
	if currPos == -1 {
		return 0, fmt.Errorf("category %q not found", categoryName)
	}

	// Find the adjacent category in the given direction
	neighborPos := -1
	neighborName := ""
	for _, o := range oracles {
		if o.Category == categoryName {
			continue
		}
		cp := o.CategoryPosition
		if direction == -1 && cp < currPos {
			if neighborPos == -1 || cp > neighborPos {
				neighborPos = cp
				neighborName = o.Category
			}
		} else if direction == 1 && cp > currPos {
			if neighborPos == -1 || cp < neighborPos {
				neighborPos = cp
				neighborName = o.Category
			}
		}
	}
	if neighborPos == -1 {
		return 0, nil // Already at boundary
	}

	if err := s.repo.SwapCategoryPositions(categoryName, neighborName, currPos, neighborPos); err != nil {
		return 0, err
	}
	return firstID, nil
}

func (s *Service) reorderOracle(oracles []*Oracle, oracleID int64, direction int) (int64, error) {
	idx := -1
	for i, o := range oracles {
		if o.ID == oracleID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return 0, fmt.Errorf("oracle %d not found", oracleID)
	}

	curr := oracles[idx]
	neighborIdx := idx + direction
	if neighborIdx < 0 || neighborIdx >= len(oracles) {
		return 0, nil // List boundary
	}
	neighbor := oracles[neighborIdx]
	if neighbor.Category != curr.Category {
		return 0, nil // Category boundary — oracles cannot cross categories
	}

	if err := s.repo.SwapPositionsInCategory(curr.ID, curr.PositionInCategory, neighbor.ID, neighbor.PositionInCategory); err != nil {
		return 0, err
	}
	return curr.ID, nil
}

// Lookup resolves an oracle reference to its weighted entry list. Implements dice.OracleLookup.
//
// Two forms are supported:
//   - "name"           — merges entries from all tables with that name (case-insensitive)
//   - "category/name"  — returns entries from the specific category's table only
func (s *Service) Lookup(name string) ([]string, bool) {
	var oracles []*Oracle

	if idx := strings.Index(name, "/"); idx != -1 {
		category := name[:idx]
		tableName := name[idx+1:]
		o, err := s.repo.GetByCategoryAndName(category, tableName)
		if err != nil {
			return nil, false
		}
		oracles = []*Oracle{o}
	} else {
		all, err := s.repo.GetAllByName(name)
		if err != nil || len(all) == 0 {
			return nil, false
		}
		oracles = all
	}

	var entries []string
	for _, o := range oracles {
		for _, line := range strings.Split(o.Content, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				entries = append(entries, line)
			}
		}
	}
	return entries, len(entries) > 0
}
