package main

import (
	"slices"
	"testing"
)

func TestConfigurationLogParsing(t *testing.T) {
	// setup

	tests := []struct {
		Index    string
		Input    string
		Expected []string
	}{
		{"01", "all", []string{LogVersionSpecialAll}},
		{"02", "none", []string{}},
		{"03", ", none ,", []string{}},
		{"04", "MS,AM", []string{LogVersionAM, LogVersionMS}},
		{"05", "MS,AM,MS,AM", []string{LogVersionAM, LogVersionMS}},
		{"06", "MS, VO AM", []string(nil)},
		{"07", "json", []string{LogVersionJSONReceiverRequest, LogVersionJSONReceiverResponse, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"08a", "jsonreq,Am", []string(nil)},
		{"08b", "jsonsend,Am", []string{LogVersionAM, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"09a", "VI,jsonresp,MA", []string(nil)},
		{"09b", "VI,jsonrecv,MA", []string(nil)},
		{"10a", "MS, Sender json Response,jsonreq", []string(nil)},
		{"10b", "MS, Sender json Response,jsonsend", []string{LogVersionMS, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"10c", "MS, Receiver json Request,jsonsend", []string{LogVersionMS, LogVersionJSONReceiverRequest, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"11", " MS, VO ,AM, OV ", []string(nil)},
		{"12", " nooone ", []string(nil)},
		{"13", " ", []string(nil)},
		{"14", "none,MS,AM", []string(nil)},
		{"15a", "json, jsonreq", []string(nil)},
		{"15b", "json, jsonsend", []string{LogVersionJSONReceiverRequest, LogVersionJSONReceiverResponse, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"16a", ",json, jsonreq", []string(nil)},
		{"16b", ",json, jsonsend", []string{LogVersionJSONReceiverRequest, LogVersionJSONReceiverResponse, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"17a", "am,json, jsonreq", []string(nil)},
		{"17b", "am,json, jsonsend", []string{LogVersionAM, LogVersionJSONReceiverRequest, LogVersionJSONReceiverResponse, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse}},
		{"18", "all, json", []string{LogVersionSpecialAll}},
	}

	for _, test := range tests {
		t.Run(test.Index, getComparerLog(test.Input, test.Expected))
	}

	// end
}

func TestConfigurationStageParsing(t *testing.T) {

	tests := []struct {
		Index    string
		Input    string
		Expected []string
	}{
		{"01", "all", []string{StageVersionSpecialAll}},
		{"02", "none", []string{}},
		{"03", ", none ,", []string{}},
		{"04", "2,1", []string{"1", "2"}},
		{"05", "1, 3,3,1", []string{"1", "3"}},
		{"06", "3, 1 1", []string(nil)},
		{"07", "last", []string{StageVersionSpecialLast}},
		{"08", ", laSt ,", []string{StageVersionSpecialLast}},
		{"09", ", last, nOne", []string(nil)},
		{"10", "all, last", []string{StageVersionSpecialAll}},
		{"11", " nooone ", []string(nil)},
		{"12", " ", []string(nil)},
		{"13", "1,1 ,1, ALL , ", []string{StageVersionSpecialAll}},
		{"14", "0,1,2,-3,4,5,6", []string(nil)},
		{"15", "1,2, 3,3,3,3,3, last", []string(nil)},
		{"16", "1,2, 3,3,3,3,3, ", []string{"1", "2", "3"}},
	}

	for _, test := range tests {
		t.Run(test.Index, getComparerStage(test.Input, test.Expected))
	}

}

func getComparerLog(input string, expected []string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Logf(`Input    : %#v`, input)

		version, err := processLogVersionsConfig(input)
		t.Logf(`Expected : %#v`, expected)
		t.Logf(`Parsed as: %#v`, version)
		if err != nil {
			t.Logf(`Error msg: %s`, err)
		}

		if !slices.Equal(version, expected) {
			t.Fail()
		}
	}
}

func getComparerStage(input string, expected []string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Logf(`Input    : %#v`, input)

		version, err := processStageVersionsConfig(input)
		t.Logf(`Expected : %#v`, expected)
		t.Logf(`Parsed as: %#v`, version)
		if err != nil {
			t.Logf(`Error msg: %s`, err)
		}

		if !slices.Equal(version, expected) {
			t.Fail()
		}
	}
}

func TestConnectionFile(t *testing.T) {
	tests := []struct {
		Index    string
		Filename string
		Expected ConnectionOptions
	}{
		{"01", "01.testdata", ConnectionOptions{}},
		{"02", "02.testdata", ConnectionOptions{Hostname: "http://yandex.loc", Username: "TESTUSER", Password: "PASSWORD"}},
		{"03", "03.testdata", ConnectionOptions{Hostname: "https://yandex.loc:50001", Username: "TEST_USER", Password: "PASS WORD"}},
		{"04", "04.testdata", ConnectionOptions{}},
		{"05", "05.testdata", ConnectionOptions{}},
		{"06", "06.testdata", ConnectionOptions{}},
		{"07", "07.testdata", ConnectionOptions{Hostname: "https://yandex.loc:50001", Username: "TESTUSER", Password: "PASSWORD"}},
		{"08", "08.testdata", ConnectionOptions{}},
		{"09", "09.testdata", ConnectionOptions{}},
		{"10", "10.testdata", ConnectionOptions{}},
		{"11", "11.testdata", ConnectionOptions{}},
		{"12", "12.testdata", ConnectionOptions{}},
		{"13", "13.testdata", ConnectionOptions{}},
		{"14", "14.testdata", ConnectionOptions{Hostname: "http://YANDEX.loc", Username: "TESTUSER", Password: "PASSWORD"}},
	}

	for _, test := range tests {
		t.Run(test.Index, getComparerConnect(test.Filename, test.Expected))
	}
}

func getComparerConnect(filename string, expected ConnectionOptions) func(t *testing.T) {
	return func(t *testing.T) {
		connect, err := GetConnectionConfig(RuntimeConfiguration{ConnectionFilepath: "testdata/connection/" + filename})
		t.Logf(`Expected : %#v`, expected)
		t.Logf(`Parsed as: %#v`, connect)
		if err != nil {
			t.Logf(`Error msg: %s`, err)
		}

		if connect != expected {
			t.Fail()
		}
	}
}

func TestGroupFlag(t *testing.T) {
	tests := []struct {
		Index    string
		Pattern  string
		Expected OutputGroup
	}{
		{"NONE", "", OutputGroupNone},
		{"NONE", " ", OutputGroupNone},
		{"NONE", "None", OutputGroupNone},
		{"NONE", " N ", OutputGroupNone},

		{"MESSAGE", "m", OutputGroupMessage},
		{"MESSAGE", "MSG", OutputGroupMessage},
		{"MESSAGE", "Messge", OutputGroupError},
		{"MESSAGE", " Message ", OutputGroupMessage},

		{"MSGVER", "MV", OutputGroupMessageVersion},
		{"MSGVER", "MsgVer", OutputGroupMessageVersion},
		{"MSGVER", "msg-ver", OutputGroupError},

		{"VERSION", "ver", OutputGroupVersion},
		{"VERSION", "vers", OutputGroupError},
	}

	for _, test := range tests {
		t.Run(test.Index, getComparerGroupFlag(test.Pattern, test.Expected))
	}

}

func getComparerGroupFlag(pattern string, expected OutputGroup) func(t *testing.T) {
	return func(t *testing.T) {
		option, err := processGroupingFlag(pattern)
		t.Logf(`Expected : %#v`, expected)
		t.Logf(`Parsed as: %#v`, option)
		if err != nil {
			t.Logf(`Error msg: %s`, err)
		}

		if option != expected {
			t.Fail()
		}
	}
}
