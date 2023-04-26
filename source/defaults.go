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

import "path/filepath"

const (
	sous_dir    = ".souschef"
	order_dir   = ".souschef/orders"
	config_path = ".souschef/config.toml"
	order_name  = "order.toml"
)

func order_path(project_dir, name string) string {
	return filepath.Join(project_dir, order_dir, name)
}

func manifest_path(project_dir, name string) string {
	return filepath.Join(project_dir, order_dir, name, order_name)
}