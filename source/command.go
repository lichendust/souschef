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
	"time"
	"bytes"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

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
			remove_file(order_path(project_dir, job.Name))
			fmt.Printf("removed job %q\n", job.Name)
		}
	}
}

func command_redo(project_dir string, args *arguments) {
	queue, ok := load_orders(project_dir, false)
	if !ok {
		return
	}

	for _, job := range queue {
		if job.Name == args.source_path {
			job.Complete = false
			job.Time     = time.Now()

			serialise_job(job, manifest_path(project_dir, job.Name))
			break
		}
	}
}

func command_targets(project_dir string, args *arguments) {
	config, ok := load_config(filepath.Join(project_dir, config_path))
	if !ok {
		return
	}

	if args.source_path != "" && args.output_path != "" {
		config.Blender_Target = append(config.Blender_Target, &Blender_Version{
			Name: args.source_path,
			Path: filepath.ToSlash(args.output_path),
		})

		buffer := bytes.Buffer{}
		buffer.Grow(512)

		if err := toml.NewEncoder(&buffer).Encode(config); err != nil {
			eprintln("failed to encode config file")
		}
		if err := os.WriteFile(filepath.Join(project_dir, config_path), buffer.Bytes(), 0777); err != nil {
			eprintln("failed to write config file")
		}
	}

	for _, t := range config.Blender_Target {
		printf("%-20s %s\n", t.Name, t.Path)
	}
}