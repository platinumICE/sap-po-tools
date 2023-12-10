package main

import (
	"fmt"
	"slices"
	"strconv"
	"sync/atomic"
)

const (
	QoS_BestEffort string = "BE"
)

func Downloader(options RuntimeConfiguration, connect ConnectionOptions, msgChannel <-chan XIAdapterMessage, versionChan chan<- XIMessageVersion) {
	defer wgDownloaders.Done()

	for msg := range msgChannel {

		if msg.QualityOfService != QoS_BestEffort && len(options.SaveStagingVersions) > 0 {
			// staged is requested and possible

			var versionsToDownload []string
			allVersions := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}
			// this is an overdoing but 5 must be enough

			if slices.Equal(options.SaveStagingVersions, []string{StageVersionSpecialAll}) {
				// ALL
				i, _ := strconv.Atoi(msg.Version)
				if i > 15 {
					// safeguard
					i = 15
				}
				versionsToDownload = allVersions[:i+1]
			} else if slices.Equal(options.SaveStagingVersions, []string{StageVersionSpecialLast}) {
				// LAST
				versionsToDownload = []string{msg.Version}
			} else {
				// all specified manually by user
				versionsToDownload = options.SaveStagingVersions
			}

			maxAvailableVersion, _ := strconv.Atoi(msg.Version)
			for _, versionName := range versionsToDownload {
				requestedVersion, _ := strconv.Atoi(versionName)

				if maxAvailableVersion < requestedVersion {
					// skip
					continue
				}

				envelop := downloadStagedVersion(connect, msg.MessageKey, versionName)

				if len(envelop) > 0 {
					versionChan <- XIMessageVersion{
						MessageInfo:    msg,
						VersionType:    VersionTypeStaged,
						MessageVersion: versionName,
						Base64Contents: envelop,
					}

					atomic.AddInt64(&statistics.NetworkBytesDownloaded, int64(len(envelop)))
				}
			}

		}

		if len(options.SaveLoggingVersions) > 0 {

			var logVersionsToDownload []string
			if slices.Equal(options.SaveLoggingVersions, []string{LogVersionSpecialAll}) {
				logVersionsToDownload = msg.LogLocations.String
			} else {
				logVersionsToDownload = options.SaveLoggingVersions
			}

			for _, versionName := range logVersionsToDownload {

				if slices.Index(msg.LogLocations.String, versionName) == -1 {
					// skip non-existant
					continue
				}

				envelop := downloadLoggedVersion(connect, msg.MessageKey, versionName)

				if len(envelop) > 0 {
					versionChan <- XIMessageVersion{
						MessageInfo:    msg,
						VersionType:    VersionTypeLogged,
						MessageVersion: versionName,
						Base64Contents: envelop,
					}

					atomic.AddInt64(&statistics.NetworkBytesDownloaded, int64(len(envelop)))
				}
			}
		}
		atomic.AddInt32(&statistics.MessagesDownloaded, 1)
	}
}

func downloadStagedVersion(connect ConnectionOptions, messageKey string, versionName string) string {

	requestTemplate := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:AdapterMessageMonitoringVi">
   <soapenv:Header/>
   <soapenv:Body>
      <urn:getMessageBytesJavaLangStringIntBoolean>
         <urn:messageKey>%s</urn:messageKey>
         <urn:version>%s</urn:version>
         <urn:archive>false</urn:archive>
      </urn:getMessageBytesJavaLangStringIntBoolean>
   </soapenv:Body>
</soapenv:Envelope>`, messageKey, versionName)

	httpResults := downloadGeneric(connect, requestTemplate)

	return httpResults.Body.GetMessageBytesJavaLangStringIntBooleanResponse.Response
}

func downloadLoggedVersion(connect ConnectionOptions, messageKey string, messageVersion string) string {

	requestTemplate := fmt.Sprintf(`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:AdapterMessageMonitoringVi">
   <soapenv:Header/>
   <soapenv:Body>
      <urn:getLoggedMessageBytes>
         <urn:messageKey>%s</urn:messageKey>
         <urn:version>%s</urn:version>
         <urn:archive>false</urn:archive>
      </urn:getLoggedMessageBytes>
   </soapenv:Body>
</soapenv:Envelope>`, messageKey, messageVersion)

	httpResults := downloadGeneric(connect, requestTemplate)

	return httpResults.Body.GetLoggedMessageBytesResponse.Response
}
