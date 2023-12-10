package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type RuntimeConfiguration struct {
	ConnectionFilepath  string
	MessageListFile     string
	MessageListFilename string
	OutputDirectory     string
	GroupOutputBy       OutputGroup
	OpenTargetDirectory bool
	ZipMode             OutputZipMode
	DownloadThreads     int
	SaveRawContent      bool
	SaveXIHeader        bool
	StatisticsOnly      bool
	SaveStagingVersions []string
	SaveLoggingVersions []string
}

type ConnectionOptions struct {
	Hostname string
	Username string
	Password string
}

type OutputZipMode string

const (
	ZipNone OutputZipMode = "none"
	ZipFile               = "file"
	ZipAll                = "all"
)

type OutputGroup string

const (
	OutputGroupNone           OutputGroup = ""
	OutputGroupError                      = "error"
	OutputGroupMessage                    = "msg"
	OutputGroupVersion                    = "version"
	OutputGroupMessageVersion             = "msgversion"
	OutputGroupVersionMessage             = "versionmsg"
)

const (
	LogVersionBI                   string = "BI"
	LogVersionVI                          = "VI"
	LogVersionMS                          = "MS"
	LogVersionAM                          = "AM"
	LogVersionVO                          = "VO"
	LogVersionJSONReceiverRequest         = "Receiver JSON Request"
	LogVersionJSONSenderRequest           = "Sender JSON Request"
	LogVersionJSONSenderResponse          = "Sender JSON Response"
	LogVersionJSONReceiverResponse        = "Receiver JSON Response"

	LogVersionSpecialAll          = "all"
	LogVersionSpecialNone         = "none"
	LogVersionSpecialJSON         = "json"
	LogVersionSpecialJSONSender   = "jsonsend"
	LogVersionSpecialJSONReceiver = "jsonrecv"
)

const (
	StageVersionSpecialAll  string = "all"
	StageVersionSpecialLast        = "last"
	StageVersionSpecialNone        = "none"
)

func ParseLaunchOptions() (RuntimeConfiguration, error) {
	options := new(RuntimeConfiguration)

	flag.StringVar(&options.ConnectionFilepath, "connection", "", "Required. Path to connection file (contains systems address, username and password)")
	messageIDs := flag.String("ids", "", "Required. Path to a list of message IDs to download, one message per line.")
	logVersions := flag.String("log", LogVersionSpecialAll,
		fmt.Sprintf(
			"Comma-separated list of log versions which must be exported. Supports standard version names (BI, MS, etc) and special values (%s, %s, %s). See details in documentation. ",
			LogVersionSpecialAll,
			LogVersionSpecialNone,
			LogVersionSpecialJSON))

	stageVersions := flag.String("stage", StageVersionSpecialAll,
		fmt.Sprintf(
			"Comma-separated list of staging version numbers (0, 1, 2, ...) which must be exported. Special values (%s, %s, %s) are acceptable. See details in documentation.",
			StageVersionSpecialAll,
			StageVersionSpecialLast,
			StageVersionSpecialNone))

	flag.BoolVar(&options.SaveXIHeader, "xiheader", false, "If specified, XI header will be saved as payload")
	flag.BoolVar(&options.SaveRawContent, "raw", false, "If specified, raw contents (multipart message format) be saved as payload")
	groupBy := flag.String("groupby", "version", "Group payloads by message ID, message version or both")
	flag.StringVar(&options.OutputDirectory, "output", "./export/", "Destination folder to save exported payloads")
	flag.BoolVar(&options.OpenTargetDirectory, "opendir", false, "Open destination folder in Explorer when download process ends")
	zipMode := flag.String("zip", "all", "Mode of compression for exported payloads. Available options are: (n)one, (f)ile, (a)ll")
	flag.IntVar(&options.DownloadThreads, "threads", 2, "Number of parallel HTTP download threads")
	flag.BoolVar(&options.StatisticsOnly, "statsonly", false, "If specified, only statistics on available message versions will be displayed. No actual download will happen.")

	//////////////

	flag.Parse()

	if messageIDs != nil {
		options.MessageListFile = *messageIDs

		fi, err := os.Lstat(options.MessageListFile)
		if err == nil && fi.IsDir() == false {
			options.MessageListFilename = fi.Name()
		}
	}

	if zipMode != nil {
		switch strings.ToLower(*zipMode) {
		case "n", "none":
			options.ZipMode = ZipNone
		case "a", "all":
			options.ZipMode = ZipAll
		case "f", "file":
			options.ZipMode = ZipFile
		default:
			return *options, fmt.Errorf("Unsupported ZIP option: [%s]", *zipMode)
		}
	}

	if logVersions != nil {
		logList, err := processLogVersionsConfig(*logVersions)
		if err != nil {
			return *options, err
		}

		options.SaveLoggingVersions = logList
	}

	if stageVersions != nil {
		stageList, err := processStageVersionsConfig(*stageVersions)
		if err != nil {
			return *options, err
		}

		options.SaveStagingVersions = stageList
	}

	if groupBy != nil {
		groupByParsed, err := processGroupingFlag(*groupBy)
		if err != nil {
			return *options, err
		}
		options.GroupOutputBy = groupByParsed
	} else {
		options.GroupOutputBy = OutputGroupNone
	}

	//////////// checks

	if options.DownloadThreads < 1 {
		return *options, fmt.Errorf("Number of download threads must be no less than 1. Value [%d] is incorrect", options.DownloadThreads)
	}

	if len(options.SaveLoggingVersions) == 0 && len(options.SaveStagingVersions) == 0 {
		return *options, fmt.Errorf("No message versions are selected for export")
	}

	return *options, nil
}

func processLogVersionsConfig(input string) ([]string, error) {
	supportedTokens := []string{LogVersionBI, LogVersionVI, LogVersionMS, LogVersionAM, LogVersionVO, LogVersionJSONReceiverRequest, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse, LogVersionJSONReceiverResponse, LogVersionSpecialAll, LogVersionSpecialNone, LogVersionSpecialJSON, LogVersionSpecialJSONSender, LogVersionSpecialJSONReceiver}
	/////////
	supportedTokensLowercase := make([]string, len(supportedTokens))

	copy(supportedTokensLowercase, supportedTokens)
	for i, _ := range supportedTokensLowercase {
		supportedTokensLowercase[i] = strings.ToLower(supportedTokensLowercase[i])
	}
	//////////
	tempList := []string{}

	pieces := strings.Split(input, ",")
	for i, _ := range pieces {
		s := strings.TrimSpace(strings.ToLower(pieces[i]))

		if s == "" {
			// skip empty
			continue
		}

		idx := slices.Index(supportedTokensLowercase, s)
		if idx == -1 {
			return nil, fmt.Errorf(`unsupported token [%s]`, pieces[i])
		}

		token := supportedTokens[idx]
		switch token {
		case LogVersionSpecialJSON:
			tempList = append(tempList, LogVersionJSONReceiverRequest, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse, LogVersionJSONReceiverResponse)

		case LogVersionSpecialJSONSender:
			tempList = append(tempList, LogVersionJSONSenderRequest, LogVersionJSONSenderResponse)

		case LogVersionSpecialJSONReceiver:
			tempList = append(tempList, LogVersionJSONReceiverRequest, LogVersionJSONReceiverResponse)

		default:
			tempList = append(tempList, token)
		}

	}
	sort.Strings(tempList)
	tempList = slices.Compact(tempList)

	// special parsings
	// nothing found
	if len(tempList) == 0 {
		return nil, fmt.Errorf(`no tokens specified`)
	}

	// single NONE
	if slices.Equal(tempList, []string{LogVersionSpecialNone}) {
		return []string{}, nil
	}

	// NONE and anything else
	if len(tempList) > 1 && slices.Contains(tempList, LogVersionSpecialNone) {
		return nil, fmt.Errorf(`cannot use "%s" with any other tokens`, LogVersionSpecialNone)
	}

	// ALL and anything else
	if len(tempList) > 1 && slices.Contains(tempList, LogVersionSpecialAll) {
		return []string{LogVersionSpecialAll}, nil
	}

	return tempList, nil
}

func processStageVersionsConfig(input string) ([]string, error) {
	supportedTokens := []string{StageVersionSpecialAll, StageVersionSpecialLast, StageVersionSpecialNone}
	/////////
	supportedTokensLowercase := make([]string, len(supportedTokens))

	copy(supportedTokensLowercase, supportedTokens)
	for i, _ := range supportedTokensLowercase {
		supportedTokensLowercase[i] = strings.ToLower(supportedTokensLowercase[i])
	}
	//////////
	tempList := []string{}

	pieces := strings.Split(input, ",")
	for i, _ := range pieces {
		s := strings.TrimSpace(strings.ToLower(pieces[i]))

		if s == "" {
			// skip empty
			continue
		}

		idx := slices.Index(supportedTokensLowercase, s)
		var token string
		if idx == -1 {
			// for numbers
			token = s
		} else {
			token = supportedTokens[idx]
		}

		switch token {
		case StageVersionSpecialNone, StageVersionSpecialLast, StageVersionSpecialAll:
			tempList = append(tempList, token)

		default:
			i, err := strconv.Atoi(token)
			if err != nil {
				return nil, fmt.Errorf(`unsupported token [%s]`, token)
			}
			if i < 0 {
				return nil, fmt.Errorf(`number is incorrect [%s]`, token)
			}
			tempList = append(tempList, token)
		}
	}

	sort.Strings(tempList)
	tempList = slices.Compact(tempList)

	// special parsings
	// nothing found
	if len(tempList) == 0 {
		return nil, fmt.Errorf(`no tokens specified`)
	}

	// single NONE
	if slices.Equal(tempList, []string{StageVersionSpecialNone}) {
		return []string{}, nil
	}

	// NONE and anything else
	if len(tempList) > 1 && slices.Contains(tempList, StageVersionSpecialNone) {
		return nil, fmt.Errorf(`cannot use "%s" with any other tokens`, StageVersionSpecialNone)
	}

	// ALL and anything else
	if len(tempList) > 1 && slices.Contains(tempList, StageVersionSpecialAll) {
		return []string{StageVersionSpecialAll}, nil
	}

	// LAST and anything else
	if len(tempList) > 1 && slices.Contains(tempList, StageVersionSpecialLast) {
		return nil, fmt.Errorf(`cannot use "%s" with any other tokens`, StageVersionSpecialLast)
	}

	return tempList, nil
}

func GetConnectionConfig(options RuntimeConfiguration) (ConnectionOptions, error) {
	contents, err := ioutil.ReadFile(options.ConnectionFilepath)
	if err != nil {
		return ConnectionOptions{}, fmt.Errorf("Configuration file [%s] not found", options.ConnectionFilepath)
	}

	lines := strings.SplitN(string(contents), "\n", 4)
	if len(lines) < 3 {
		// not enough lines
		return ConnectionOptions{}, fmt.Errorf("Connection file [%s] is incorrect", options.ConnectionFilepath)
	}

	if len(lines) == 4 && strings.TrimSpace(lines[3]) != "" {
		// too many lines with content
		return ConnectionOptions{}, fmt.Errorf("Connection file [%s] is incorrect", options.ConnectionFilepath)
	}

	connect := ConnectionOptions{
		Hostname: strings.TrimSpace(lines[0]),
		Username: strings.TrimSpace(lines[1]),
		Password: strings.TrimSpace(lines[2]),
	}

	parsedURL := new(url.URL)
	parsedURL, err = url.Parse(connect.Hostname)
	if err != nil {
		return ConnectionOptions{}, fmt.Errorf("Configuration file [%s] contains invalid URL specification [%s]", options.ConnectionFilepath, connect.Hostname)
	}

	switch parsedURL.Scheme {
	case "http", "https":
		//noop
	default:
		return ConnectionOptions{}, fmt.Errorf("Configuration file [%s] contains invalid URL specification [%s]", options.ConnectionFilepath, connect.Hostname)
	}

	connect.Hostname = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	if connect.Username == "" || connect.Password == "" {
		// no login or password
		return ConnectionOptions{}, fmt.Errorf("Connection file [%s] is incorrect, login or password are not provided", options.ConnectionFilepath)
	}

	return connect, nil
}

func processGroupingFlag(input string) (OutputGroup, error) {
	switch strings.TrimSpace(strings.ToLower(input)) {
	case "", "n", "none":
		return OutputGroupNone, nil
	case "m", "msg", "message":
		return OutputGroupMessage, nil
	case "v", "ver", "version":
		return OutputGroupVersion, nil
	case "vm", "vermsg", "versionmessage":
		return OutputGroupVersionMessage, nil
	case "mv", "msgver", "messageversion":
		return OutputGroupMessageVersion, nil
	default:
		return OutputGroupError, fmt.Errorf(`Group option [%s] is unknown`, input)
	}
}
