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

const config_file = `# the version to use by default when creating
# a new job
default_target = "3.5"

# these are example versions that may not
# match your system. please add, remove or
# update paths that are relevant to you
[[blender_target]]
name = "2.93"
path = "~/software/blender_2.93/blender"

[[blender_target]]
name = "3.5"
path = "~/software/blender_3.5/blender"

[[blender_target]]
name = "canary"
path = "~/dev/buildbot/blender`