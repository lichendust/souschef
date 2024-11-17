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
import "path/filepath"

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
}

func command_list(config *Config) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
		printf("No orders found!\n")
		return
	}

	printf("\n")

	index := 0
	for _, order := range queue {
		if order.Complete {
			printf("✓  ")
		} else {
			index += 1
			printf("%-3d", index)
		}

		printf(apply_color("[$1%s$0] %s\n"), order.Name, filepath.Base(order.Source_Path))

		printf("   Using:        %s\n",       order.Blender_Target)
		printf("   Frame Range:  %d -> %d\n", order.Start_Frame,  order.End_Frame)
		printf("   Resolution:   %d x %d\n",  order.Resolution_X, order.Resolution_Y)

		printf("   Placeholders: %s\n", format_fallback_bool(order.Use_Placeholders))
		printf("   Overwriting:  %s\n", format_fallback_bool(order.Overwrite))

		if order.Output_Path == "." {
			printf("   Output Path:  %s\n", SET_BY_FILE)
		} else {
			printf("   Output Path:  %s\n", order.Output_Path)
		}

		printf("\n")
	}
}

func command_clean(config *Config, args *Arguments) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
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
		printf("0/%d orders are eligible for deletion. Use --hard to force\n", len(queue))
		return
	}

	count := 0

	for _, order := range queue {
		if args.hard_clean || order.Complete {
			remove_file(order_path(config.project_dir, order.Name))
			printf(apply_color("[$1%s$0] removed!\n"), order.Name)
			count += 1
		}
	}

	printf("%d/%d orders removed\n", count, len(queue))
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

	// @todo if something else removes its lock file,
	// Sous Chef should probably stop that job for safety reasons
}

func command_delete(config *Config, args *Arguments) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	for _, order := range queue {
		if order.Name == args.source_path {
			remove_file(order_path(config.project_dir, order.Name))
			break
		}
	}
}

func command_targets(config *Config, args *Arguments) {
	if len(config.Blender_Target) == 0 {
		printf("No Blender targets in config.toml\n")
		return
	}

	print("Target Name          Blender Path\n")

	for _, t := range config.Blender_Target {
		if !file_exists(t.Path) {
			printf(apply_color("$1%-20s %s\n$0"), t.Name, t.Path)
		} else {
			printf("%-20s %s\n", t.Name, t.Path)
		}
	}
}
