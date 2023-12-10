package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"
)

func prepareFileWriter(options RuntimeConfiguration, connect ConnectionOptions) error {
	if options.StatisticsOnly {
		// nothing to do here
		return nil
	}

	if options.OutputDirectory == "" {
		return errors.New("Destination directory is not specified")
	}

	url, err := url.Parse(connect.Hostname)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot parse hostname: %s", err))
	}

	dt := time.Now().Format("20060102150405")

	path := fmt.Sprintf("%s/%s/%s/", options.OutputDirectory, url.Hostname(), dt)

	err = os.MkdirAll(path, 0750)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot create output directory [%s]: %s", path, err))
	}

	err = os.Chdir(path)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot change current directory to [%s]: %s", path, err))
	}

	return nil
}

func FileWriter(options RuntimeConfiguration, version <-chan XIMessagePayloads) {
	defer wgWriters.Done()

	switch options.ZipMode {
	case ZipNone:
		FileWriterModeNone(options, version)
	case ZipFile:
		FileWriterModeFile(options, version)
	case ZipAll:
		FileWriterModeAll(options, version)
	default:
		panic("not supported ZipMode: " + options.ZipMode)
	}

}

func openTargetDirectory(options RuntimeConfiguration) {

	if options.OpenTargetDirectory == true {
		cmd := exec.Command("explorer", ".")
		_ = cmd.Run()
	}

}

func createPath(folder string) string {
	if folder == "" {
		return "."
	}

	err := os.MkdirAll(folder, 0750)
	if err != nil {
		fmt.Printf("Cannot create folder [%s] for export\n", folder)
		os.Exit(7)
	}
	return folder
}
