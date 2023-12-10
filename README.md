# SAP PO Tools

Collection of SAP PO tools which are created and used for every day use during SAP PO integration scenario monitoring and maintenance. All tools are tried and tested in live environments.

## SAP PO Tools -- Downloader

This tool allows to download SAP PO message payloads and attachments given list of message IDs and connection details. The tool allows you to download:

- any staged message version (async only)
- any logged message version (sync and async) including request and response JSON

This tool is able to download messages in parallel, by default two HTTP threads are used. Number of threads is configurable.

More details are available in [separate README page](https://github.com/platinumICE/sap-po-tools/downloader).
