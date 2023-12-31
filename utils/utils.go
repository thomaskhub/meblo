package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// define ffmpef error codes to access it globaly
const (
	FFmpegErrorEinval = -22
	FFmpegErrorEof    = -541478725
	FFmpegErrorAgain  = -11
	// add more error codes as needed
)

func IsFilePath(path string) bool {
	// Regular expressions for Linux, Windows, and Mac file paths
	linuxRegex := `^\.*/`
	windowsRegex := `^[a-zA-Z]:\\.*$` //TODO: not tested yet as we run only on linux
	macRegex := `^/Volumes/.*$`       //TODO: not tested yet as we run only on linux

	// Check if the string matches any of the regular expressions
	isLinuxPath, _ := regexp.MatchString(linuxRegex, path)
	isWindowsPath, _ := regexp.MatchString(windowsRegex, path)
	isMacPath, _ := regexp.MatchString(macRegex, path)

	// Return true if the string matches any of the regular expressions
	return isLinuxPath || isWindowsPath || isMacPath
}

// ConvertToAbsolutePath takes a relative path as input and returns the corresponding absolute path.
//
// Parameters:
//
//	path (string): The relative path to be converted.
//
// Returns:
//
//	string: The absolute path.
//	error: The error, if any, occurred during the conversion.
func ConvertToAbsolutePath(path string, workDir string) string {
	// Get the current working directory

	//check if the path is relative it could be a linux windows or mac path strnig
	// Check if the path is a relative path
	if !filepath.IsAbs(path) {

		// Join the current working directory and the relative path
		absolutePath := filepath.Join(workDir, path)

		// Return the absolute path
		return absolutePath
	}

	return path
}

// create function GetCurrentWorkingDir
// GetCurrentWorkingDir returns the current working directory.
func GetCurrentWorkingDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(dir), nil
}

// GetOutputFormat get the output format based on the outStr provided
// if outStr is an rtmpUrl return flv, if it is an dash url return dash, if it is a file return mpegts
func GetOutputFormat(path string) string {
	// check if the outString is an rtmp url if so set outFormat to OutputModeRtmp
	if strings.HasPrefix(path, "rtmp://") {
		return OutputModeRtmp
	}
	if strings.HasSuffix(path, ".mpd") {
		return OutputModeDash
	}

	if strings.HasSuffix(path, ".ts") {
		return OutputModeMpegts
	}

	return ""
}
