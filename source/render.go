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

import "os"
import "fmt"
import "time"
import "bufio"
import "os/exec"
import "strings"
import "unicode"
import "path/filepath"

type Blender_Error uint8
const (
	ALL_GOOD Blender_Error = iota
	NO_MEMORY
	NO_VIDEO_MEMORY
	FILESYSTEM_ERROR
	PYTHON_FAIL
	RENDERER_CRASH
	RENDERER_NOT_SUPPORTED
	RENDERER_KERNEL_FAIL
	GPU_NOT_SUPPORTED
	EXCEPTION_ACCESS_VIOLATION
)

func (e Blender_Error) String() string {
	switch e {
	case NO_MEMORY:
		return "out of RAM"
	case NO_VIDEO_MEMORY:
		return "out of VRAM"
	case FILESYSTEM_ERROR:
		return "failed to read data from filesystem"
	case PYTHON_FAIL:
		return "python failed to initialise"
	case RENDERER_CRASH:
		return "renderer crashed"
	case RENDERER_NOT_SUPPORTED:
		return "renderer not supported"
	case RENDERER_KERNEL_FAIL:
		return "CUDA kernel failed to compile"
	case GPU_NOT_SUPPORTED:
		return "graphics card not supported"
	case EXCEPTION_ACCESS_VIOLATION:
		// this one is non-specific, so passing it on directly
		// will help people find the real answer faster by
		// searching themselves — it could be drivers, display
		// properties and (notably) some weird Windows quirks
		return "EXCEPTION_ACCESS_VIOLATION"
	}
	return "" // never happens
}

func check_progress(order *Order, input string) string {
	buffer := strings.Builder{}

	if strings.HasPrefix(input, "Fra:") {
		buffer.Grow(64)

		for i, c := range input {
			if unicode.IsSpace(c) {
				the_frame := input[4:i]

				percentage, ok := parse_uint(the_frame)
				if ok {
					percentage = uint(float64(percentage - order.Start_Frame) / float64(order.frame_count) * 100)
				}

				buffer.WriteString(fmt.Sprintf("| %d%% %s ", percentage, the_frame))

				if strings.Contains(input, "Compositing") {
					buffer.WriteString(apply_color("$1composite$0 "))
				} else {
					buffer.WriteString(apply_color("$1render$0 "))
				}
				break
			}
		}

		{
			index := strings.Index(input, "Time:")

			if index > -1 {
				the_time := strings.TrimSpace(input[index + 5:])

				for i, c := range the_time {
					if unicode.IsSpace(c) {
						the_time = the_time[:i]
						break
					}
				}

				buffer.WriteString(the_time)
			}
		}
	}

	return buffer.String()
}

// @todo we're probably missing a lot of errors nowadays
func check_errors(input string) Blender_Error {
	switch true {
	case strings.Contains(input, "std::bad_alloc"):
		return NO_MEMORY
	case strings.Contains(input, "alloc returns null"):
		return NO_MEMORY
	case strings.Contains(input, "CUDA kernel compilation failed"):
		return RENDERER_KERNEL_FAIL
	case strings.Contains(input, "CUDA device supported only with compute capability"):
		return GPU_NOT_SUPPORTED
	case strings.Contains(input, "CUDA error"):
		return RENDERER_CRASH
	case strings.Contains(input, "terminate called after throwing an instance of 'boost::filesystem::filesystem_error'"):
		return FILESYSTEM_ERROR
	case strings.Contains(input, "Fatal Python error: Py_Initialize"):
		return PYTHON_FAIL
	case strings.Contains(input, "Warning: Cycles is not enabled!"):
		return RENDERER_NOT_SUPPORTED
	case strings.Contains(input, "not available for scene"):
		return RENDERER_NOT_SUPPORTED
	case strings.Contains(input, "EXCEPTION_ACCESS_VIOLATION"):
		return EXCEPTION_ACCESS_VIOLATION
	}
	return ALL_GOOD
}

func command_render(config *Config, args *Arguments) {
	queue, ok := load_orders(config.project_dir, false)
	if !ok {
		return
	}

	if len(queue) == 0 {
		printf("No orders to render!\n")
		return
	}

	for len(queue) > 0 {
		the_order := queue[0]

		if the_order.Complete {
			queue = queue[1:]
			continue
		}

		if the_order.lock != "" && the_order.lock != config.own_hostname {
			queue = queue[1:]
			continue
		}

		lock_file := lock_path(config.project_dir, the_order.Name)
		write_file(lock_file, config.own_hostname)

		did_run := run_order(config, the_order)
		if !did_run {
			println("Failed!")
			queue = queue[1:]
			continue
		}

		os.Remove(lock_file)

		did_save := save_order(the_order, manifest_path(config.project_dir, the_order.Name))
		if !did_save {
			print("\n") // preserve the error emitted by save_order
			queue = queue[1:]
			continue
		}

		queue = queue[1:]
	}
}

func run_order(config *Config, order *Order) bool {
	blender_path, got_path := get_blender_path(config, order.Blender_Target)
	if !got_path {
		return false
	}

	target := filepath.Join(config.project_dir, order.Target_Path)

	// output    := filepath.Join(project_dir, order.Output_Path)      "-o"
	// format, _ := get_image_types(filepath.Ext(order.Output_Path))   "-F"

	the_command := exec.Command(blender_path, "-b", target, "--python-expr", inject(config.project_dir, order), "-a")

	stdout, err := the_command.StdoutPipe()
	if err != nil {
		return false
	}

	err = the_command.Start()
	if err != nil {
		return false
	}

	go func() {
		time.Sleep(time.Second)

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			message := check_progress(order, line)
			printf(apply_color(RESET_LINE + "[$1%s$0] %s %s"), order.Name, filepath.Base(order.Target_Path), message)

			program_state := check_errors(line)
			if program_state != ALL_GOOD {
				eprintln("error", program_state.String())
				break
			}
		}
	}()

	err = the_command.Wait()

	if err != nil {
		return false
	}

	order.Complete = true
	printf(" ✓\n")

	return true
}

const PATH_REWRITER = `
import os
import bpy

from os.path import *

def is_abs(path):
    if path.startswith("//"):
        path = "." + path[1:]
    path = bpy.path.native_pathsep(path)
    return isabs(path)

def abs_path(path):
    return normpath(bpy.path.abspath(path))

def dirname(path):
    return path[:-len(basename(path))]

output_path = "%s"
source_path = abs_path(bpy.context.scene.render.filepath)

common_paths = []

for node in bpy.context.scene.node_tree.nodes:
    if node.mute or "ignore" in node.label.lower():
        continue
    if node.type == 'OUTPUT_FILE':
        if is_abs(node.base_path):
            continue
        node_path = abs_path(node.base_path)
        common_paths.append(commonpath([source_path, node_path]))

if len(common_paths) > 0:
    shortest_common = min(common_paths, key=len)

    for node in bpy.context.scene.node_tree.nodes:
        if node.mute or "ignore" in node.label.lower():
            continue
        if node.type == 'OUTPUT_FILE':
            if is_abs(node.base_path):
                continue
            node_path = abs_path(node.base_path)[len(shortest_common) + 1:]
            node.base_path = join(output_path, node_path) + os.sep

    if len(common_paths) > 0:
        bpy.context.scene.render.filepath = join(output_path, source_path[len(shortest_common) + 1:])

else:
	bpy.context.scene.render.filepath = output_path
`

func inject(project_dir string, order *Order) string {
	buffer := new(strings.Builder)
	buffer.Grow(512)

	buffer.WriteString("import bpy\n")

	if order.Output_Path != "." {
		path := filepath.ToSlash(filepath.Join(project_dir, order.Output_Path))
		buffer.WriteString(fmt.Sprintf(PATH_REWRITER, path))
	}

	// auto-tiling for Blender 3+
	buffer.WriteString("bpy.context.scene.cycles.use_auto_tile = (bpy.app.version[0] < 3)\n")

	buffer.WriteString("bpy.context.scene.render.use_render_cache = True\n")

	buffer.WriteString(fmt.Sprintf("bpy.context.scene.frame_start = %d\n", order.Start_Frame))
	buffer.WriteString(fmt.Sprintf("bpy.context.scene.frame_end   = %d\n", order.End_Frame))

	if order.Resolution_X > 0 && order.Resolution_Y > 0 {
		buffer.WriteString(fmt.Sprintf("bpy.context.scene.render.resolution_x = %d\n", order.Resolution_X))
		buffer.WriteString(fmt.Sprintf("bpy.context.scene.render.resolution_y = %d\n", order.Resolution_Y))
		buffer.WriteString("bpy.context.scene.render.resolution_percentage = 100\n")
	}

	if order.Use_Placeholders != UNSPECIFIED {
		buffer.WriteString("bpy.context.scene.render.use_placeholder = ")
		if order.Use_Placeholders == YES {
			buffer.WriteString("True\n")
		} else {
			buffer.WriteString("False\n")
		}
	}

	if order.Overwrite != UNSPECIFIED {
		buffer.WriteString("bpy.context.scene.render.use_overwrite = ")
		if order.Overwrite == YES {
			buffer.WriteString("True\n")
		} else {
			buffer.WriteString("False\n")
		}
	}

	return buffer.String()
}

/*func get_image_types(ext string) (string, bool) {
	switch strings.ToLower(ext) {
	case ".bmp":            return "BMP",      true
	case ".cin", ".cineon": return "CINEON",   true
	case ".hdr":            return "HDR",      true
	case ".iris":           return "IRIS",     true
	case ".iriz":           return "IRIZ",     true
	case ".jp2":            return "JP2",      true
	case ".jpg", ".jpeg":   return "JPEG",     true
	case ".exr":            return "OPEN_EXR", true
	case ".png":            return "PNG",      true
	case ".tga":            return "TGA",      true
	case ".tif", ".tiff":   return "TIFF",     true
	case ".webp":           return "WEBP",     true
	}

	// "AVIJPEG"
	// "AVIRAW"
	// "DDS"
	// "DPX"
	// "MPEG"
	// "OPEN_EXR_MULTILAYER"
	// "RAWTGA"

	return "", false
}*/
