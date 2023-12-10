package main

type XIEnvelop struct {
	Body Body `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type Body struct {
	GetMessageBytesJavaLangStringIntBooleanResponse XIgetMessageBytesJavaLangStringIntBooleanResponse `xml:"urn:AdapterMessageMonitoringVi getMessageBytesJavaLangStringIntBooleanResponse"`
	GetLoggedMessageBytesResponse                   XIgetLoggedMessageBytesResponse                   `xml:"urn:AdapterMessageMonitoringVi getLoggedMessageBytesResponse"`
	GetMessageListResponse                          XIgetMessageListResponse                          `xml:"urn:AdapterMessageMonitoringVi getMessageListResponse"`
	Manifest                                        XIManifest                                        `xml:"http://sap.com/xi/XI/Message/30 Manifest"`
}

type XIgetLoggedMessageBytesResponse struct {
	Response string `xml:"urn:AdapterMessageMonitoringVi Response"`
}

type XIgetMessageBytesJavaLangStringIntBooleanResponse struct {
	Response string `xml:"urn:AdapterMessageMonitoringVi Response"`
}

type XIgetMessageListResponse struct {
	Response struct {
		List struct {
			AdapterFrameworkData []XIAdapterMessage `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws AdapterFrameworkData"`
		} `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws list"`
	} `xml:"urn:AdapterMessageMonitoringVi Response"`
}

type XIAdapterMessage struct {
	Direction        string         `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws direction"`
	MessageID        string         `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws messageID"`
	MessageKey       string         `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws messageKey"`
	QualityOfService string         `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws qualityOfService"`
	Version          string         `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws version"`
	LogLocations     XILogLocations `xml:"urn:com.sap.aii.mdt.server.adapterframework.ws logLocations"`
}

type XILogLocations struct {
	String []string `xml:"urn:java/lang String"`
}

type XIManifest struct {
	Payload []struct {
		Href        string `xml:"http://www.w3.org/1999/xlink href,attr"`
		Name        string `xml:"http://sap.com/xi/XI/Message/30 Name"`
		Description string `xml:"http://sap.com/xi/XI/Message/30 Description"`
		Type        string `xml:"http://sap.com/xi/XI/Message/30 Type"`
	} `xml:"http://sap.com/xi/XI/Message/30 Payload"`
}
