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