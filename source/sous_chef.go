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
import "os/exec"
import "strings"
import "path/filepath"
import "github.com/BurntSushi/toml"

const VERSION = "v0.2.0"
const PROGRAM = "Sous Chef " + VERSION

const SOUS_DIR      = ".souschef"
const ORDER_DIR     = SOUS_DIR + "/orders"
const CONFIG_PATH   = SOUS_DIR + "/config.toml"
const MANIFEST_NAME = "order.toml"
const LOCK_NAME     = "lock.txt"

const (
	COMMAND_ORDER uint8 = iota
	COMMAND_VERSION
	COMMAND_HELP
	COMMAND_INIT
	COMMAND_LIST
	COMMAND_CLEAN
	COMMAND_RENDER
	COMMAND_REDO
	COMMAND_DELETE
	COMMAND_TARGET
)

type Arguments struct {
	command    uint8
	hard_clean bool

	replace_id string

	bank_order       bool
	start_frame      uint
	end_frame        uint
	resolution_x     uint
	resolution_y     uint
	overwrite        uint8
	use_placeholders uint8
	source_path      string
	output_path      string
	blender_target   string

	is_bat_installed bool
}

type Config struct {
	project_dir  string
	own_hostname string

	Default_Target string             `toml:"default_target"`
	Blender_Target []*Blender_Version `toml:"target"`
}

type Blender_Version struct {
	Name string `toml:"name"`
	Path string `toml:"path"`
}

func main() {
	args, ok := get_arguments()
	if !ok {
		return
	}

	switch args.command {
	case COMMAND_INIT:
		command_init()
		return

	case COMMAND_HELP:
		command_help()
		return

	case COMMAND_VERSION:
		println(PROGRAM)
		return
	}

	config, ok := load_config()
	if !ok {
		eprintf("Cannot locate .souschef project!\n")
		return
	}

	switch args.command {
	case COMMAND_LIST:
		command_list(config)

	case COMMAND_CLEAN:
		command_clean(config, args)

	case COMMAND_ORDER:
		command_order(config, args)

	case COMMAND_RENDER:
		command_render(config, args)

	case COMMAND_REDO:
		command_redo(config, args)

	case COMMAND_DELETE:
		command_delete(config, args)

	case COMMAND_TARGET:
		command_targets(config, args)
	}
}

func load_config() (*Config, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, false
	}

	found := false
	for {
		if file_exists(filepath.Join(cwd, SOUS_DIR)) {
			found = true
			break
		}

		l := len(cwd) - len(filepath.Base(cwd)) - 1
		if l < 0 {
			break
		}

		println(cwd[:1])

		cwd = cwd[:l]
		if len(cwd) == 0 {
			break
		}
	}

	if !found {
		return nil, false
	}

	conf_path := filepath.Join(cwd, OS_CONFIG_PATH)
	if !file_exists(conf_path) {
		conf_path = filepath.Join(cwd, CONFIG_PATH)
	}

	blob, ok := load_file(conf_path)
	if !ok {
		return nil, false
	}

	data := new(Config)

	_, err = toml.Decode(blob, data)
	if err != nil {
		return nil, false
	}

	data.own_hostname = hostname()
	data.project_dir  = cwd

	return data, true
}

func get_blender_path(config *Config, t string) (string, bool) {
	blender_path := ""
	found_path   := false

	if t == "" {
		t = config.Default_Target
	}

	for _, target := range config.Blender_Target {
		if target.Name == t {
			found_path = true
			blender_path = target.Path
			break
		}
	}

	if !found_path {
		for _, target := range config.Blender_Target {
			if target.Name == config.Default_Target {
				found_path = true
				blender_path = target.Path
				break
			}
		}

		if !found_path {
			eprintf(apply_color("Target $1%q$0 not in config.toml\n"), t)
			return "", false
		}
	}

	return blender_path, true
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

func get_arguments() (*Arguments, bool) {
	args := os.Args[1:]
	conf := new(Arguments)

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

			case "delete":
				conf.command = COMMAND_DELETE
				args = args[1:]
				continue

			case "targets":
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
			conf.bank_order = true
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

		case "overwrite", "o":
			counter++
			b = strings.ToLower(b)
			if b == "yes" {
				conf.overwrite = YES
			} else if b == "no" {
				conf.overwrite = NO
			}
			continue

		case "placeholders", "p":
			counter++
			b = strings.ToLower(b)
			if b == "yes" {
				conf.use_placeholders = YES
			} else if b == "no" {
				conf.use_placeholders = NO
			}
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
			eprintf("Arguments: %q flag is unknown\n", a)
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
			eprintf("Arguments: too many path arguments\n")
			has_errors = true
		}

		patharg++
	}

	if conf.command == COMMAND_ORDER && conf.source_path == "" {
		conf.command = COMMAND_HELP
		has_errors = true
	}

	// check for the existence of BAT
	{
		_, err := exec.LookPath("bat")
		if err == nil {
			conf.is_bat_installed = true
		}
	}

	return conf, !has_errors
}
