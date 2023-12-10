package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"slices"
	"sort"
	"strings"
)

// this will also handle ID file to messageID channel convertion
type XIMessageInfo struct {
	MessageKey       string
	MessageID        string
	Direction        string
	QualityOfService string
}

func prepareMessageList(options RuntimeConfiguration) ([]string, error) {
	if options.MessageListFile == "" {
		return nil, fmt.Errorf("Message list is not specified")
	}

	contents, err := ioutil.ReadFile(options.MessageListFile)
	if err != nil {
		return nil, fmt.Errorf("Message list file %s not found", options.MessageListFile)
	}

	readlines := strings.Split(string(contents), "\n")
	lines := make([]string, 0)
	for _, l := range readlines {
		l = strings.ToLower(strings.TrimSpace(l))

		// GUID must match formats:
		// aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
		//  -- OR --
		// aaaaaaaabbbbccccddddeeeeeeeeeeee
		matched, _ := regexp.MatchString(`^([0-9a-f]{32}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`, l)
		if !matched {
			continue
		}

		lines = append(lines, l)
	}

	// sort and remove duplicates
	sort.Strings(lines)
	lines = slices.Compact(lines)

	// stats
	statistics.MessagesInFile = int32(len(lines))

	return lines, nil
}

func searchMessages(options RuntimeConfiguration, connect ConnectionOptions, idList []string, msgChannel chan<- XIAdapterMessage) error {
	response, err := search(connect, idList)
	if err != nil {
		return err
	}

	// stats
	statistics.MessagesFound = int32(len(response.Response.List.AdapterFrameworkData))

	go func(response XIgetMessageListResponse, msgChannel chan<- XIAdapterMessage) {
		for _, msg := range response.Response.List.AdapterFrameworkData {
			msgChannel <- msg
		}

		close(msgChannel)
	}(response, msgChannel)

	return nil
}

func search(connect ConnectionOptions, IDList []string) (XIgetMessageListResponse, error) {

	IDListFormatted := ""
	for _, id := range IDList {
		IDListFormatted += fmt.Sprintf("<lang:String>%s</lang:String>", id)
	}

	requestTemplate := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:AdapterMessageMonitoringVi" xmlns:urn1="urn:com.sap.aii.mdt.server.adapterframework.ws" xmlns:urn2="urn:com.sap.aii.mdt.api.data" xmlns:lang="urn:java/lang">
	   <soapenv:Header/>
	   <soapenv:Body>
	      <urn:getMessageList>
         <urn:filter>
           <urn1:archive>false</urn1:archive>
            <urn1:dateType>0</urn1:dateType>
            <urn1:messageIDs>%s</urn1:messageIDs>
            <urn1:nodeId>0</urn1:nodeId>
            <urn1:onlyFaultyMessages>false</urn1:onlyFaultyMessages>
           <urn1:retries>0</urn1:retries>
            <urn1:retryInterval>0</urn1:retryInterval>
           <urn1:timesFailed>0</urn1:timesFailed>
           <urn1:wasEdited>false</urn1:wasEdited>
            <urn1:returnLogLocations>true</urn1:returnLogLocations>
            <urn1:onlyLogLocationsWithPayload>true</urn1:onlyLogLocationsWithPayload>
         </urn:filter>
	         <urn:maxMessages>%d</urn:maxMessages>
	      </urn:getMessageList>
	   </soapenv:Body>
	</soapenv:Envelope>`, IDListFormatted, 10000)
	// maxMessages is set to sensible default
	// TODO: add processing for continueation

	httpResults := downloadGeneric(connect, requestTemplate)

	if len(httpResults.Body.GetMessageListResponse.Response.List.AdapterFrameworkData) == 0 {
		return XIgetMessageListResponse{}, errors.New("no messages found in target system")
	}

	return httpResults.Body.GetMessageListResponse, nil

}
