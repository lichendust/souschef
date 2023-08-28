/*
	Sous Chef
	Copyright (C) 2022-2023 Harley Denham

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

import "os"
import "fmt"
import "io/fs"
import "errors"
import "strings"
import "strconv"
import "unicode/utf8"
import "path/filepath"
import "github.com/mattn/go-isatty"

func order_path(project_dir, name string) string {
	return filepath.Join(project_dir, ORDER_DIR, name)
}

func manifest_path(project_dir, name string) string {
	return filepath.Join(project_dir, ORDER_DIR, name, MANIFEST_NAME)
}

func lock_path(project_dir, name string) string {
	return filepath.Join(project_dir, ORDER_DIR, name, LOCK_NAME)
}

func file_exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func load_file(path string) (string, bool) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	return string(content), true
}

func make_directory(path string) bool {
	err := os.MkdirAll(path, os.ModeDir | os.ModePerm)

	if err != nil {
		eprintf("failed to create directory %q\n", path)
		return false
	}

	return true
}

func write_file(path, content string) bool {
	err := os.WriteFile(path, []byte(content), os.ModePerm)
	if err != nil {
		eprintf("failed to write file %q\n", path)
		return false
	}

	return true
}

func remove_file(path string) bool {
	err := os.RemoveAll(path)
	if err != nil {
		eprintf("failed to delete file %q\n", path)
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

var running_in_term = false

func init() {
	running_in_term = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func print(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
}

func println(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
	os.Stdout.WriteString("\n")
}

func printf(format string, guff ...any) {
	fmt.Fprintf(os.Stdout, format, guff...)
}

/*func eprint(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stderr.WriteString(w)
		if i < l {
			os.Stderr.WriteString(" ")
		}
	}
}*/

func eprintln(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stderr.WriteString(w)
		if i < l {
			os.Stderr.WriteString(" ")
		}
	}
	os.Stderr.WriteString("\n")
}

func eprintf(format string, guff ...any) {
	fmt.Fprintf(os.Stderr, format, guff...)
	os.Stderr.WriteString("\n")
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		eprintf("\n    failed to get hostname! is this computer okay?\n\n")
		return ""
	}
	return h
}

const ANSI_RESET = "\033[0m"
const ANSI_COLOR = "\033[91m"
const RESET_LINE = "\033[2K\r"

func apply_color(input string) string {
	buffer := strings.Builder{}
	buffer.Grow(len(input) + 128)

	last_rune := 'x'

	for {
		if len(input) == 0 {
			break
		}

		r, w := utf8.DecodeRuneInString(input)
		input = input[w:]

		if r == '$' {
			last_rune = r
			continue
		}

		if last_rune == '$' {
			last_rune = 'x'

			if r == '0' || r == '1' {
				if !running_in_term {
					continue
				} else if r == '0' {
					buffer.WriteString(ANSI_RESET)
				} else {
					buffer.WriteString(ANSI_COLOR)
				}
			} else {
				buffer.WriteRune('$')
				buffer.WriteRune(r)
			}

			continue
		}

		last_rune = r
		buffer.WriteRune(r)
	}

	return buffer.String()
}