package main

import (
	"os"
	"errors"
	"io/ioutil"
	"path/filepath"
)

func find_directory(input string) (string, bool) {
	{
		// path := filepath.Join(input, jobs_path)

		path := input

		if file_exists(path) {
			return path, true
		}
	}

	// go up one directory and try again
	base := filepath.Base(input)
	input = input[:len(input) - len(base) - 1]

	// break if we've reached the top
	if len(input) == 0 {
		return "", false
	}

	return find_directory(input)
}

func file_exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func load_file(path string) (string, bool) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return "", false
	}

	return string(content), true
}