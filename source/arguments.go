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
	"os"
	"fmt"
	"strings"
)

const (
	COMMAND_ORDER uint8 = iota
	COMMAND_VERSION
	COMMAND_HELP
	COMMAND_INIT
	COMMAND_LIST
	COMMAND_CLEAN
	COMMAND_RENDER
	COMMAND_REDO
	COMMAND_TARGET
)

type arguments struct {
	command     uint8
	watch_files bool
	hard_clean  bool

	replace_id  string

	// order details
	bank_job    bool
	start_frame uint
	end_frame   uint

	resolution_x uint
	resolution_y uint

	source_path    string
	output_path    string
	blender_target string
}

func preset_res_table(arg string) (uint, uint) {
	switch strings.ToLower(arg) {
	case "uhd":
		return 3840, 2160
	case "hd":
		return 1920, 1080
	case "dcp4k":
		return 4096, 1716
	case "dcp2k":
		return 2048, 858
	}
	return 0, 0
}

// extracts arguments in the array as
// either --bool or --name <data>
func pull_argument(args []string) (string, string) {
	if len(args) == 0 {
		return "", ""
	}

	if len(args[0]) >= 1 {
		n := 0

		for _, c := range args[0] {
			if c != '-' {
				break
			}
			n++
		}

		a := args[0]

		if n > 0 {
			a = a[n:]
		} else {
			return "", ""
		}

		if len(args[1:]) >= 1 {
			b := args[1]

			if len(b) > 0 && b[0] != '-' {
				return a, b
			}
		}

		return a, ""
	}

	return "", ""
}

func get_arguments() (*arguments, bool) {
	conf       := &arguments {}
	args       := os.Args[1:]
	counter    := 0
	patharg    := 0
	has_errors := false

	for {
		args = args[counter:]

		if len(args) == 0 {
			break
		}

		counter = 0

		if len(args) > 0 {
			switch args[0] {
			case "init":
				conf.command = COMMAND_INIT
				args = args[1:]
				continue

			case "order":
				conf.command = COMMAND_ORDER
				args = args[1:]
				continue

			case "list":
				conf.command = COMMAND_LIST
				args = args[1:]
				continue

			case "render":
				conf.command = COMMAND_RENDER
				args = args[1:]
				continue

			case "clean":
				conf.command = COMMAND_CLEAN
				args = args[1:]
				continue

			case "redo":
				conf.command = COMMAND_REDO
				args = args[1:]
				continue

			case "target":
				conf.command = COMMAND_TARGET
				args = args[1:]
				continue

			case "help":
				conf.command = COMMAND_HELP
				return conf, true // exit immediately

			case "version":
				conf.command = COMMAND_VERSION
				return conf, true // exit immediately
			}
		}

		a, b := pull_argument(args[counter:])

		counter++

		switch a {
		case "":
		case "cache", "c":
			conf.bank_job = true
			continue

		case "watch", "w":
			conf.watch_files = true
			continue

		case "target", "t":
			conf.blender_target = b
			continue

		case "hard":
			conf.hard_clean = true
			continue

		case "replace":
			counter++
			conf.replace_id = b
			continue

		case "resolution", "r":
			counter++
			part := strings.SplitN(b, "x", 2)

			switch len(part) {
			case 1:
				conf.resolution_x, conf.resolution_y = preset_res_table(part[0])

				if conf.resolution_x == 0 {
					eprintf("unknown preset %q\n", part[0])
				}
			case 2:
				if x, ok := parse_uint(part[0]); ok {
					conf.resolution_x = x
				}
				if y, ok := parse_uint(part[1]); ok {
					conf.resolution_y = y
				}
			}
			continue

		case "frame", "f":
			counter++
			part := strings.SplitN(b, ":", 2)

			switch len(part) {
			case 1:
				if x, ok := parse_uint(part[0]); ok {
					conf.end_frame = x
				}
				conf.start_frame = 1
			case 2:
				if x, ok := parse_uint(part[0]); ok {
					conf.start_frame = x
				}
				if x, ok := parse_uint(part[1]); ok {
					conf.end_frame = x
				}
			}
			continue

		case "version":
			conf.command = COMMAND_VERSION
			return conf, true

		case "help", "h":
			// psychological failsafe —
			// the user is most likely
			// to try "--help" or "-h" first
			conf.command = COMMAND_HELP
			return conf, true

		default:
			fmt.Fprintf(os.Stderr, "args: %q flag is unknown\n", a)
			has_errors = true

			if b != "" {
				counter++
			}
		}

		switch patharg {
		case 0:
			conf.source_path = args[0]
		case 1:
			conf.output_path = args[0]
		default:
			fmt.Fprintln(os.Stderr, "args: too many path arguments")
			has_errors = true
		}

		patharg++
	}

	if conf.command == COMMAND_ORDER && conf.source_path == "" {
		fmt.Fprintln(os.Stderr, "no file specified")
		has_errors = true
	}

	return conf, !has_errors
}