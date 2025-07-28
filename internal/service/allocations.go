package service

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

func GetAllocations() (
	[]AllocationRule,
	error,
) {
	if allocStore == nil {
		return nil, ErrNoAllocStore
	}

	rules, err := allocStore.GetAllocations()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return rules, nil
}

func SetAllocations(
	rules []AllocationRule,
) error {
	if allocStore == nil {
		return ErrNoAllocStore
	}

	// ensure total percentage adds to 100
	total := 0
	for _, r := range rules {
		total += r.Percentage
	}
	if total != 100 {
		return ErrInvalidAlloc
	}

	err := allocStore.SetAllocations(rules)
	if err != nil {
		return DatabaseError{err}
	}

	return nil
}
