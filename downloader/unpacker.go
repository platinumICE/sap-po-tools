package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"strings"
	"sync/atomic"
)

type XIMessagePayloads struct {
	MessageID string
	VersionID string
	Folder    string
	Parts     []XIPayload
}

type XIPayload struct {
	Filename string
	Contents []byte
}

type XIMessageVersion struct {
	MessageInfo    XIAdapterMessage
	VersionType    VersionType
	MessageVersion string
	Base64Contents string
}

type VersionType string

const (
	VersionTypeLogged VersionType = "LOG"
	VersionTypeStaged             = "STAGE"
)

const MultipartRelated string = "multipart/related"

func Unpacker(options RuntimeConfiguration, versionChan <-chan XIMessageVersion, payloadChan chan<- XIMessagePayloads) {
	defer wgUnpackers.Done()

	for entry := range versionChan {
		payloadChan <- UnpackPartsBase64(options, entry)
	}
}

func generateFilename(filename string) string {
	// clean the name from non-FS symbols, just in case
	filename_clean := strings.Map(func(s rune) rune {
		if strings.IndexRune(`<>:"/\|?*`, s) == -1 {
			return s
		} else {
			return '_'
		}
	}, filename)
	return filename_clean
}

func generateFilenamePrefix(options RuntimeConfiguration, entry XIMessageVersion) (string, string) {

	pathtemplate := ""
	filenameprefixtemplate := ""

	switch options.GroupOutputBy {
	// %[1]s:  MessageID
	// %[2]s:  VersionType
	// %[3]s:  MessageVersion
	case OutputGroupMessage:
		pathtemplate = "%[1]s"
		filenameprefixtemplate = "%[2]s.%[3]s."

	case OutputGroupVersion:
		pathtemplate = "%[2]s.%[3]s"
		filenameprefixtemplate = "%[1]s."

	case OutputGroupMessageVersion:
		pathtemplate = "%[1]s/%[2]s.%[3]s"
		filenameprefixtemplate = ""

	case OutputGroupVersionMessage:
		pathtemplate = "%[2]s.%[3]s/%[1]s"
		filenameprefixtemplate = ""

	case OutputGroupNone:
		pathtemplate = ""
		filenameprefixtemplate = "%[1]s.%[2]s.%[3]s."

	default:
		pathtemplate = ""
		filenameprefixtemplate = "%[1]s.%[2]s.%[3]s."
	}

	path, filenameprefix := "", ""
	if pathtemplate != "" {
		path = fmt.Sprintf(pathtemplate, entry.MessageInfo.MessageID, entry.VersionType, entry.MessageVersion)
	}
	if filenameprefixtemplate != "" {
		filenameprefix = fmt.Sprintf(filenameprefixtemplate, entry.MessageInfo.MessageID, entry.VersionType, entry.MessageVersion)
	}

	return path, filenameprefix

}

func UnpackPartsBase64(options RuntimeConfiguration, entry XIMessageVersion) XIMessagePayloads {
	data, err := base64.StdEncoding.DecodeString(entry.Base64Contents)
	if err != nil {
		fmt.Printf("Error unpacking base64 contents for Message Key [%s], skipping...\n", entry.MessageInfo.MessageKey)
		return XIMessagePayloads{}
	}
	return UnpackParts(options, entry, data)
}

func UnpackParts(options RuntimeConfiguration, entry XIMessageVersion, data []byte) XIMessagePayloads {

	path, filenameprefix := generateFilenamePrefix(options, entry)

	payloads := XIMessagePayloads{}
	payloads.MessageID = entry.MessageInfo.MessageID
	payloads.VersionID = fmt.Sprintf("%s.%s", entry.VersionType, entry.MessageVersion)
	payloads.Folder = path

	if options.SaveRawContent {

		payloads.Parts = append(payloads.Parts, XIPayload{
			Filename: generateFilename(filenameprefix + "RAW"),
			Contents: bytes.Clone(data), // copy of bytes since we are making adjustments later on
		})

		// RAW is not considered a payload, so do not count it
		// atomic.AddInt64(&statistics.PayloadSize, int64(len(data)))
		// atomic.AddInt32(&statistics.PayloadsExtracted, 1)
	}

	// fix for SAP mistake in mime/multipart
	data = bytes.Replace(data, []byte("\n\r"), []byte("\r\n"), 1)

	////
	bufReader := bytes.NewReader(data)
	reader := bufio.NewReader(bufReader)
	headerAttributes := make(map[string]string)
	// only single-line header declarations are supported for now
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			break
		}
		parts := strings.SplitN(string(line), ":", 2)
		key := strings.ToLower(parts[0])
		value := parts[1]

		headerAttributes[key] = value
	}

	val, ok := headerAttributes["content-type"]
	if !ok {
		fmt.Printf("Cannot decode payload for [%s], no content-type attribute, skipping...\n", entry.MessageInfo.MessageID)
		return payloads
	}

	mediatype, params, _ := mime.ParseMediaType(val)
	if !strings.HasPrefix(mediatype, MultipartRelated) {
		fmt.Printf("Unsupported content-type for [%s] = [%s], should be [%s]. Skipping...\n", entry.MessageInfo.MessageID, mediatype, MultipartRelated)
		return payloads
	}

	mr := multipart.NewReader(reader, params["boundary"])
	xiHeaderContentID := params["start"]
	xiMessageHeader := XIManifest{}

	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Printf("Error at NextPart [%s]. skipping....\n", err)
			break
		}

		partData, err := io.ReadAll(p)
		if err != nil {
			fmt.Printf("Error reading Part [%s]. skipping....\n", err)
			break
		}

		partContentID := p.Header.Get("Content-Id")
		partContentType := p.Header.Get("Content-Type")
		partFilename := p.FileName()

		if partContentID == xiHeaderContentID {
			xiMessageHeader = processXIHeader(partData)
			if options.SaveXIHeader == false {
				// skipping header
				continue
			}

			partFilename = "XIHEADER.xml" // constant name
		}

		if partFilename == "" {
			partFilename = getPartNameByContentID(xiMessageHeader, partContentID, partContentType)
		}

		filename := generateFilename(filenameprefix + partFilename)
		if filename == "" {
			fmt.Printf("Cannot generate filename for [%s], skipping...\n", entry.MessageInfo.MessageID)
			continue
		}

		payloadPart := XIPayload{
			Filename: filename,
			Contents: partData,
		}

		payloads.Parts = append(payloads.Parts, payloadPart)

		atomic.AddInt32(&statistics.PayloadsExtracted, 1)
		atomic.AddInt64(&statistics.PayloadSize, int64(len(partData)))
	}

	processDuplicateFilenames(&payloads)
	return payloads
}

func processXIHeader(content []byte) XIManifest {
	result := new(XIEnvelop)
	err := xml.Unmarshal(content, &result)
	if err != nil {
		fmt.Printf(`cannot process XI header, need revision`)
		return XIManifest{}
	}

	return result.Body.Manifest
}

func getPartNameByContentID(xiMessageHeader XIManifest, contentID string, contentType string) string {

	contentID = "cid:" + strings.Trim(contentID, "<>")
	probableName := ""

	for _, payload := range xiMessageHeader.Payload {
		if payload.Href != contentID {
			continue
		}

		probableName = payload.Name
		if probableName == "" {
			// fall back to message type
			probableName = payload.Type
		}

		if probableName == "" {
			// fall back to content ID
			probableName = strings.Trim(contentID, "<>")
		}

		break
	}

	return probableName
}

func processDuplicateFilenames(payloads *XIMessagePayloads) {

	names := make(map[string]bool)

	for i, _ := range payloads.Parts {
		name := payloads.Parts[i].Filename
		postfix := 1
		newname := name

		for {
			newname = name
			if postfix > 1 {
				newname = fmt.Sprintf("%s_%d", newname, postfix)
			}

			_, found := names[newname]
			if !found {
				names[newname] = true
				break
			} else {
				postfix += 1
			}
		}

		if name != newname {
			payloads.Parts[i].Filename = newname
		}
	}
}
