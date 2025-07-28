package service

// AllocationRule describes how incoming funds are distributed
// across different ledgers.
type AllocationRule struct {
	ID         string `json:"id"`
	LedgerName string `json:"ledger"`
	Percentage int    `json:"percentage"`
}

type AllocationsStore interface {
	GetAllocations() ([]AllocationRule, error)
	SetAllocations([]AllocationRule) error
}

var allocStore AllocationsStore

func SetAllocationsStore(s AllocationsStore) {
	allocStore = s
}

// GetAllocations returns the current allocation ruleset.
func GetAllocations() ([]AllocationRule, error) {
	if allocStore == nil {
		return nil, ErrNoAllocStore
	}
	rules, err := allocStore.GetAllocations()
	if err != nil {
		return nil, DatabaseError{err}
	}
	return rules, nil
}

// SetAllocations replaces the current allocation ruleset after validating
// that the provided percentages sum to 100.
func SetAllocations(rules []AllocationRule) error {
	if allocStore == nil {
		return ErrNoAllocStore
	}
	total := 0
	for _, r := range rules {
		total += r.Percentage
	}
	if total != 100 {
		return ErrInvalidAlloc
	}
	if err := allocStore.SetAllocations(rules); err != nil {
		return DatabaseError{err}
	}
	return nil
}
