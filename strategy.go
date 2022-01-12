package concurrencyx

import "sync"

type denyStrategyTable struct {
	DenyStrategyManagementInterface
	table sync.Map
}

func (t *denyStrategyTable) AddStrategy(s Seconds, threshold DenyThreshold) error {
	if s <= 0 || s > secondsLimit {
		return ErrorSecondsFiledTooLarge
	}
	if threshold <= 0 {
		return ErrorThresholdTooSmall
	}
	t.table.Store(s, threshold)
	return nil
}

func (t *denyStrategyTable) DelStrategy(s Seconds) {
	t.table.Delete(s)
}

func (t *denyStrategyTable) OutPut() *map[Seconds]DenyThreshold {
	res := make(map[Seconds]DenyThreshold)
	t.table.Range(func(key, value interface{}) bool {
		res[key.(Seconds)] = value.(DenyThreshold)
		return true
	})
	return &res
}

func createDenyStrategyTable(denyStrategy map[Seconds]DenyThreshold) (*denyStrategyTable, error) {
	table := denyStrategyTable{table: sync.Map{}}

	for s, t := range denyStrategy {
		if err := (&table).AddStrategy(s, t); err != nil {
			return nil, err
		}
	}

	return &table, nil
}
