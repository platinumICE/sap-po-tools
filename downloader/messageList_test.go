package main

import (
	"fmt"
	"testing"
)

func TestMessageFileList(t *testing.T) {

	tests := []struct {
		Index         string
		Filename      string
		ExpectedCount int
	}{
		{"01", "list1.ok.testdata", 5},
		{"02", "list2.empty.testdata", 0},
		{"03", "list3.mixed.testdata", 5},
		{"04", "list4.abap.testdata", 5},
		{"05", "list5.broken.testdata", 0},
		{"06", "list6.sloppy.testdata", 4},
		{"07", "list7.duplicates.testdata", 3},
	}

	for _, test := range tests {
		t.Run(test.Index, getMessageListTester(test.Filename, test.ExpectedCount))
	}
}

func getMessageListTester(filename string, expected int) func(t *testing.T) {
	return func(t *testing.T) {
		options := RuntimeConfiguration{
			MessageListFile: fmt.Sprintf("./testdata/ids/%s", filename),
		}

		idList, err := prepareMessageList(options)
		t.Logf(`ID count: %d out of expected %d`, len(idList), expected)
		if err != nil {
			t.Errorf(`Error: %s`, err)
		}

		if len(idList) != expected {
			t.FailNow()
		}

	}

}
