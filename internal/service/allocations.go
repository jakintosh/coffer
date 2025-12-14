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

func (s *Service) GetAllocations() (
	[]AllocationRule,
	error,
) {
	if s == nil || s.Allocations == nil {
		return nil, ErrNoAllocStore
	}

	rules, err := s.Allocations.GetAllocations()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return rules, nil
}

func (s *Service) SetAllocations(
	rules []AllocationRule,
) error {
	if s == nil || s.Allocations == nil {
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

	err := s.Allocations.SetAllocations(rules)
	if err != nil {
		return DatabaseError{err}
	}

	return nil
}
