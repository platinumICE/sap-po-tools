# SAP PO Tools -- Downloader
## Purpose

This tool allows to download staged and logged versions of messages from SAP PO (Java Stack) using only standard supported APIs (AdapterMessageMonitoringVi service). Tool performs the following steps:

1. parses message ID list (from plain text file)
2. connects to remote SAP PO system using connection details (from connection file)
3. performs a search query on SAP PO given message IDs
4. requests available staged and logged versions (based on availability and user request)
5. parses XI messages into components (payloads)
6. saves payloads to local file system (in ZIP format)

Optionally this tool will save raw multipart message and XI header (as-is).

Each download attempt will create a folder *\<output>/\<hostname>/\<timestamp>/*. Further grouping of payloads is configurable with [-groupby](#groupby) option. 

Files will be renamed (suffix will be added) if name collisions should occur. Also some characters in filename may be replaced by underscore (\_) if they are not valid for use in filesystem.

## Usage and command-line parameters

Tool can be run with following command-line parameters:

	-connection string
          Required. Path to connection file (contains systems address, username and password)
 	-ids string
          Required. Path to a list of message IDs to download, one message per line. See detailed explanation below.
 	-log string
          Comma-separated list of log versions which must be exported. Supports standard version names (BI, MS, etc) and special values (all, none, json). (default "all") See detailed explanation below. 
	-stage string
	      Comma-separated list of staging version numbers (0, 1, 2, ...) which must be exported. Special values (all, last, none) are acceptable. (default "all") See detailed explanation below. 
	-xiheader
	      If specified, XI header will be saved as payload
	-raw
	      If specified, raw contents (multipart message format) be saved as payload
	-groupby string
          Group payloads by message ID, message version or both (default "version"). See detailed explanation below.
	-output string
	      Destination folder to save exported payloads (default "./export/")
	-opendir
	      Open destination folder in Explorer when download process ends
	-zip string
	      Mode of compression for exported payloads. Available options are: (n)one, (f)ile, (a)ll (default "all"). See detailed explanation below.
	-threads int
	      Number of parallel HTTP download threads (default 2)
	-statsonly
	      If specified, only statistics on available message versions will be displayed. No actual download will happen.
	-nocomment
		  If specified, no text comment will be added to ZIP file (applies to -zip all)

## Message ID list file format

Message ID list must be presented as plain text file with message IDs, one per line. IDs should be specied either formats:
- in UUID format (*XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX*)
- raw format (*XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX*)

Latter format is used by SXI_MONITOR for example. Whitespaces and empty lines are ignored. Any line which does not comform to specified formats will be ignored.

Example file:

	Message IDs
	e1f49308-b5d9-1ede-a1e6-a84cac19a8d2
	e1f49308-b5d9-1ede-a1e6-a84d3503c8d2
	e1f49308-b5d9-1ede-a1e6-a8744657c8d3
	a325a530-8910-11ee-9eeb-00000c9d89ea
	13dd6480-8944-11ee-badc-00000c9d89ea


## Connection file format

Connection file must be presented as plain text file with connection details in the following format:

- system URL
- username
- password

Example file:

	https://example.com/
	JOHN.SMITH
	$ecretPassw0rd

URL may link to any page on the target server. Only scheme (HTTP/HTTPS), hostname and port number are considered by the tool. 

## -log Option

Option allows to specify which Log versions of the messages must be exported if available at target SAP PO server. Tool accepts comma-separated list of Log versions which will be requested from SAP PO. Available options to specify (not case-sensative) are listed below. They map to corresponding Log versions SAP PO used.

- **BI**
- **VI**
- **MS**
- **AM**
- **VO**
- **Receiver JSON Request**
- **Sender JSON Request**
- **Sender JSON Response**
- **Receiver JSON Response**

Some special options are available for conveniance. All options below are not case-sensitive.

- **all** — will request all available versions
- **json** — maps to "*Receiver JSON Request, Sender JSON Request, Sender JSON Response, Receiver JSON Response*"
- **jsonsend** — maps to "*Sender JSON Request, Sender JSON Response*"
- **jsonrecv** — maps to "*Receiver JSON Request, Receiver JSON Response*"
- **none** — will not download any Log versions

Specifying **none** with anything else will result in error. Specifying **all** will ignore any other option except **none**.

## -stage Option

Option is not available for messages with *Best Effort* delivery semantics. Option allows to specify which Stage versions of the messages must be exported if available at target SAP PO server. Tool accepts comma-separated list of Stage versions which will be requested from SAP PO. List of stage versions (represented as integer numbers starting from zero) may be specified, and they correspond to Stage version number seen in SAP PO.

Special options listed below are available. Options are not case-sensative.

- **all**  — will request all available versions
- **last** — maps to last Stage version available (that is — with highest version number)
- **none** — will not download any Stage versions

Specifying **none** with anything else will result in error. 
Specifying **all** will ignore any other option.
Specifying **last** with anything else (except **all**) will result in error. 

## -groupby Option

Specifies the folder layout inside output folder. Available options are (not case-sensative):

	n, none:
		Files with named {MESSAGEID}.{MESSAGEVERSION}.{PAYLOADNAME}. No subfolders will be created
	m, msg, message:
		Files with named {MESSAGEVERSION}.{PAYLOADNAME}. Subfolders {MESSAGEID} will be created.
	v, ver, version:
		Files with named {MESSAGEID}.{PAYLOADNAME}. Subfolders {MESSAGEVERSION} will be created.
	vm, vervsg, versionmessage:
		Files with named {PAYLOADNAME}. Subfolders {MESSAGEVERSION}/{MESSAGEID} will be created.
	mv, msgver, messageversion:
		Files with named {PAYLOADNAME}. Subfolders {MESSAGEID}/{MESSAGEVERSION} will be created.

Default value is **version**.

## -zip Option

Specified if export should be compressed or not. Available options are (not case-sensative):

	n, none:
		All files will be written to disk without compression.
	f, file:
		All files will be written to disk but each individual file will be compressed with GZIP. Resulting filenames will have ".gz" suffix in their name.
	a, all:
		All downloaded files will be placed inside one ZIP archive. Subfolder structure (see -groupby) will be preserved.


Default value is **all**. If ZIP file is used for output, comment will be added in the format:

	Source      : <hostname>
	Extracted on: <current datetime>




