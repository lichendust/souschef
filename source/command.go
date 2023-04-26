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
)

func command_help() {
	println(PROGRAM)

	args := os.Args[2:]
	if len(args) == 0 {
		print(apply_color(help("help")))
		return
	}

	print(apply_color(help(args[0])))
}

func command_init() {
	if !make_directory(order_dir) {
		return
	}
	if !write_file(config_path, config_file) {
		return
	}
	println("initialised Sous Chef project")
}

func command_list(project_dir string) {
	queue, ok := load_orders(project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
		println("no orders found!")
		return
	}

	for i, job := range queue {
		print_order(i + 1, job)
	}
}

func command_clean(project_dir string, args *arguments) {
	queue, ok := load_orders(project_dir, false)

	if !ok {
		return
	}

	for _, job := range queue {
		if args.hard_clean || job.Complete {
			remove_file(order_path(project_dir, job.Name.word))
			fmt.Printf("removed job %q\n", job.Name)
		}
	}
}