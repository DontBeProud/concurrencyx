package concurrencyx

import "testing"

func TestStrategy(t *testing.T) {
	_, err := createDenyStrategyTable(map[Seconds]DenyThreshold{secondsLimit + 1: 100})
	if err == nil {
		t.Error("bug found in seconds check module")
		return
	}

	_, err = createDenyStrategyTable(map[Seconds]DenyThreshold{0: 1})
	if err == nil {
		t.Error("bug found in seconds check module")
		return
	}

	_, err = createDenyStrategyTable(map[Seconds]DenyThreshold{1: 0})
	if err == nil {
		t.Error("bug found in threshold check module")
		return
	}

	table, err := createDenyStrategyTable(map[Seconds]DenyThreshold{
		1:   1,
		100: 2,
	})
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = table.AddStrategy(200, 3)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = table.AddStrategy(200, 4)
	if err != nil {
		t.Error(err.Error())
		return
	}

	table.DelStrategy(100)

	o := table.OutPut()
	for key, value := range *o {
		if !((key == 1 && value == 1) || (key == 200 && value == 4)) {
			t.Error("bug found")
		}
	}

}
