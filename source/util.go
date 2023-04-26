/*
	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"io/fs"
	"io/ioutil"

	"os"
	"fmt"
	"errors"
	"strconv"
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
	return input[:len(input) - len(filepath.Base(input)) - 1]
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

func make_directory(path string) bool {
	err := os.MkdirAll(path, os.ModeDir)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create directory %q\n", path)
		return false
	}

	return true
}

func write_file(path, content string) bool {
	err := ioutil.WriteFile(path, []byte(content), 0777)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write file %q\n", path)
		return false
	}

	return true
}

func remove_file(path string) bool {
	err := os.RemoveAll(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete file %q\n", path)
		return false
	}

	return true
}

func dir_size(root string) (float64, bool) {
	total := int64(0)

	err := filepath.WalkDir(root, func(path string, file fs.DirEntry, err error) error {
		if !file.IsDir() {
			info, err := file.Info()

			if err != nil {
				panic(err)
			}

			total += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, false
	}

	return float64(total) / 1048576, true // size in megabytes
}

func parse_uint(str string) (uint, bool) {
	u, err := strconv.ParseUint(str, 0, 32)
	if err != nil {
		return 0, false
	}

	return uint(u), true
}