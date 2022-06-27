package main

import (
	"os"
	"fmt"
	"github.com/BurntSushi/toml"
)

type config struct {
	Default_Target hash
	Blender_Target []*blender_version
}

type blender_version struct {
	Name hash
	Path string
}

func load_config(path string) (*config, bool) {
	blob, ok := load_file(path)

	if !ok {
		fmt.Fprintln(os.Stderr, "failed to load config")
		return nil, false
	}

	data := config {}

	{
		_, err := toml.Decode(blob, &data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to parse config")
			return nil, false
		}
	}

	return &data, true
}