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
import "time"
import "bytes"
import "path/filepath"
import "github.com/BurntSushi/toml"

func command_help() {
	println(PROGRAM)

	args := os.Args[2:]
	if len(args) == 0 {
		println(apply_color(help("help")))
		return
	}

	println(apply_color(help(args[0])))
}

func command_init() {
	if !make_directory(ORDER_DIR) {
		return
	}
	if !write_file(CONFIG_PATH, config_file) {
		return
	}
	printf("\n    initialised Sous Chef project\n\n")
}

func command_list(config *Config) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
		printf("\n    no orders found!\n\n")
		return
	}

	print("\n")
	for i, order := range queue {
		print_order(i + 1, order)
	}
	print("\n")
}

func command_clean(config *Config, args *Arguments) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
		printf("\n    project is already clean!\n\n")
		return
	}

	can_remove_any := false
	for _, order := range queue {
		if args.hard_clean || order.Complete {
			can_remove_any = true
			break
		}
	}

	if !can_remove_any {
		printf("\n    0/%d orders are eligible for deletion. use --hard to force.\n\n", len(queue))
		return
	}

	count := 0

	for _, order := range queue {
		if args.hard_clean || order.Complete {
			remove_file(order_path(config.project_dir, order.Name))
			printf(apply_color("\n    [$1%s$0] removed"), order.Name)
			count += 1
		}
	}

	printf("\n\n    %d/%d orders removed.\n\n", count, len(queue))
}

func command_redo(config *Config, args *Arguments) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	for _, order := range queue {
		if order.Name == args.source_path {
			order.Complete = false
			order.Time     = time.Now()

			save_order(order, manifest_path(config.project_dir, order.Name))
			os.Remove(lock_path(config.project_dir, order.Name))
			break
		}
	}
}

func command_targets(config *Config, args *Arguments) {
	if args.source_path != "" && args.output_path != "" {
		name := args.source_path
		path := filepath.ToSlash(args.output_path)

		for _, c := range config.Blender_Target {
			if c.Name == name && c.Path == path {
				return
			}
		}

		config.Blender_Target = append(config.Blender_Target, &Blender_Version{
			Name: args.source_path,
			Path: filepath.ToSlash(args.output_path),
		})

		buffer := new(bytes.Buffer)
		buffer.Grow(512)

		if err := toml.NewEncoder(buffer).Encode(config); err != nil {
			eprintln("\n    failed to encode config file.")
		}
		if err := os.WriteFile(filepath.Join(config.project_dir, CONFIG_PATH), buffer.Bytes(), os.ModePerm); err != nil {
			eprintln("\n    failed to write config file.")
		}
	}

	if len(config.Blender_Target) == 0 {
		printf("\n    no Blender targets in config\n\n")
		return
	}

	print("\n    NAME                 FILEPATH\n\n")

	for _, t := range config.Blender_Target {
		if !file_exists(t.Path) {
			print(apply_color("[$1!$0] "))
		} else {
			print("    ")
		}
		printf("%-20s %s\n", t.Name, t.Path)
	}

	print("\n")
}