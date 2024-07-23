package main

import (
	"testing"
	"time"
)

type checkTimeRangeTest struct {
	begin, end, searched, result string
}

var testCaseCheckTimeRange = []checkTimeRangeTest{
	{"2023-12-30T19:08:18+07:00", "2024-01-04T15:27:03+07:00", "2024-01-01T12:00:03+07:00", "equal"},
	{"2023-12-29T20:33:51+07:00", "2024-01-16T14:58:21+07:00", "2023-12-27T08:38:16+07:00", "left"},
	{"2024-02-08T04:36:49+07:00", "2024-02-26T03:09:29+07:00", "2024-03-04T10:49:12+07:00", "right"},
}

func TestCheckTimeRange(t *testing.T) {
	for _, testCase := range testCaseCheckTimeRange {
		beginDateStr := testCase.begin
		endDateStr := testCase.end
		searchedDateStr := testCase.searched

		want := testCase.result

		beginDate, err := time.Parse(time.RFC3339, beginDateStr)
		if err != nil {
			t.Errorf("invalid begin date time: %s", err)
			return
		}

		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			t.Errorf("invalid end date time: %s", err)
			return
		}

		searchedDate, err := time.Parse(time.RFC3339, searchedDateStr)
		if err != nil {
			t.Errorf("invalid searched date time: %s", err)
			return
		}

		result := checkTimeRange(beginDate, endDate, searchedDate)

		if result != want {
			t.Fatalf(`checkTimeRange(%s, %s, %s) returns %q, expected should be: %q`, beginDateStr, endDateStr, searchedDateStr, result, want)
		}
	}
}
