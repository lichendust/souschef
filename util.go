package main

import (
	"os"
	"fmt"
	"errors"
	"strconv"
	"io/ioutil"
	"path/filepath"
)

func find_project_dir() (string, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to get working directory")
		return "", false
	}

	path, ok := recurse_dirs(cwd)
	if !ok {
		fmt.Fprintln(os.Stderr, "not a Sous Chef project!")
		return "", false
	}

	return path, true
}

func recurse_dirs(input string) (string, bool) {
	{
		path := filepath.Join(input, sous_dir)

		if file_exists(path) {
			return input, true
		}
	}

	// go up one directory and try again
	input = slice_base(input)

	// break if we've reached the top
	if len(input) == 0 {
		return "", false
	}

	return recurse_dirs(input)
}

func slice_base(input string) string {
	base := filepath.Base(input)
	return input[:len(input) - len(base) - 1]
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

func parse_uint(str string) (uint, bool) {
	u, err := strconv.ParseUint(str, 0, 32)
	if err != nil {
		return 0, false
	}
	return uint(u), true
}