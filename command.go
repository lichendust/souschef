package main

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
)

func command_help() {
	args := os.Args[1:]

	if len(args) <= 1 {
		fmt.Println(apply_color(comm_help))
		return
	}

	switch strings.ToLower(args[0]) {
	case "render":
		fmt.Println(apply_color(comm_render))
	}
}

func command_init() {
	if !make_directory(order_dir) {
		return
	}
	if !write_file(config_path, config_file) {
		return
	}
	fmt.Println("initialised Sous Chef project")
}

func command_list(project_dir string) {
	queue, ok := load_orders(project_dir, false)

	if !ok {
		return
	}

	if len(queue) == 0 {
		fmt.Println("no orders found!")
		return
	}

	for i, job := range queue {
		complete := ""

		if job.Complete {
			complete = "complete!"
		}

		fmt.Println(i + 1, job.Name, " ", filepath.Base(job.Source_Path), " ", job.Start_Frame, job.End_Frame, complete)
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