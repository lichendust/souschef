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

const OS_CONFIG_PATH = SOUS_DIR + "/config_macos.toml"

const config_file = `# the version to use by default when creating
# a new order
default_target = "4.2"

# these are example versions that may not
# match your system. please add, remove or
# update paths that are relevant to you
[[target]]
name = "3.0"
path = "/Applications/Blender 3.0.app/Contents/MacOS/blender"

[[target]]
name = "4.2"
path = "/Applications/Blender 4.2.app/Contents/MacOS/blender"

[[target]]
name = "canary"
path = "/Volumes/Development/buildbot/blender"`
