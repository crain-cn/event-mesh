package events

import (
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	filterLists := FilterList{
		"cronjob-controller",
		"job-controller",
	}
	notFilterLists := FilterList{}
	eventFilter := NewEventFilter(time.Now(), filterLists, notFilterLists)
	expected1 := true
	filter1Result := eventFilter.filterComponent("cronjob-controller")
	if filter1Result != expected1 {
		t.Errorf("\nexpected:\n%v\ngot:\n%v", expected1, filter1Result)
	}
	expected2 := true
	filter2Result := eventFilter.filterComponent("job-controller")
	if filter2Result != expected2 {
		t.Errorf("\nexpected:\n%v\ngot:\n%v", expected2, filter2Result)
	}
	expected3 := false
	filter3Result := eventFilter.filterComponent("cronjob1-controller")
	if filter3Result != expected3 {
		t.Errorf("\nexpected:\n%v\ngot:\n%v", expected3, filter3Result)
	}
}

func TestKillingFilter(t *testing.T) {
	filterLists := FilterList{
		"SuccessfulCreate",
	}
	notFilterLists := FilterList{
		"Killing",
	}

	eventFilter := NewEventFilter(time.Now(), filterLists, notFilterLists)
	expected1 := false
	filter1Result := eventFilter.filterReason("Killing")
	if filter1Result != expected1 {
		t.Errorf("\nexpected:\n%v\ngot:\n%v", expected1, filter1Result)
	}

	expected2 := true
	filter2Result := eventFilter.filterReason("SuccessfulCreate")
	if filter2Result != expected2 {
		t.Errorf("\nexpected:\n%v\ngot:\n%v", expected1, filter1Result)
	}

}
