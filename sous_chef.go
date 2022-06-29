package main

import (
	"os"
	"fmt"
)

const title = "Sous Chef 0.1.0RC"

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
		fmt.Println(title)
		command_help()
		return

	case COMMAND_VERSION:
		fmt.Println(title)
		return
	}

	project_dir, ok := find_project_dir()

	if !ok {
		fmt.Fprintln(os.Stderr, "not a Sous Chef project!")
		return
	}

	switch args.command {
	case COMMAND_LIST:
		command_list(project_dir)
		return

	case COMMAND_CLEAN:
		command_clean(project_dir, args)
		return

	case COMMAND_ORDER:
		command_order(project_dir, args)
		return

	case COMMAND_RENDER:
		command_render(project_dir, args)
		return
	}
}