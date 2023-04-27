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

import "github.com/BurntSushi/toml"

type config struct {
	Default_Target string
	Blender_Target []*Blender_Version
}

type Blender_Version struct {
	Name string
	Path string
}

func load_config(path string) (*config, bool) {
	blob, ok := load_file(path)
	if !ok {
		eprintln("failed to load config")
		return nil, false
	}

	data := config{}

	{
		_, err := toml.Decode(blob, &data)
		if err != nil {
			eprintln("failed to parse config")
			return nil, false
		}
	}

	return &data, true
}