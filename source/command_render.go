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
	"os/exec"

	"fmt"
	"time"
	"bufio"
	"strings"
	"unicode"
	"path/filepath"
)

func command_render(project_dir string, args *arguments) {
	config, ok := load_config(filepath.Join(project_dir, config_path))
	if !ok {
		return
	}

	queue, ok := load_orders(project_dir, false)

	if !ok {
		return
	}

	if len(queue) == 0 {
		fmt.Println("no orders to render!")
		return
	}

	for len(queue) > 0 {
		the_job := queue[0]

		if the_job.Complete {
			queue = queue[1:]
			continue
		}

		fmt.Printf("[%s] %s\n", strings.ToUpper(the_job.Name.word), filepath.Base(the_job.Target_Path))

		{
			ok := run_job(config, the_job, project_dir)
			if !ok {
				fmt.Println("failed!")
				queue = queue[1:]
				continue
			}
		}
		{
			ok := serialise_job(the_job, manifest_path(project_dir, the_job.Name.word))
			if !ok {
				fmt.Printf("\n")   // preserve the error emitted by serialise_job
				queue = queue[1:]
				continue
			}
		}

		fmt.Println("\033[2K\rcomplete!")
		queue = queue[1:]
	}
}

func run_job(config *config, job *Job, project_dir string) bool {
	blender_path := ""
	found_path   := false

	for _, target := range config.Blender_Target {
		if target.Name.uint32 == job.Blender_Target.uint32 {
			found_path = true
			blender_path = target.Path
			break
		}
	}

	if !found_path {
		fmt.Fprintln(os.Stderr, "specified blender target not found in config.toml")
		return false
	}

	target := filepath.Join(project_dir, job.Target_Path)

	// output    := filepath.Join(project_dir, job.Output_Path)
	// format, _ := get_image_types(filepath.Ext(job.Output_Path))

	// "-o" output
	// "-F" format

	cmd := exec.Command(blender_path, "-b", target, "--python-expr", injected_expression(project_dir, job), "-a")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}

	err = cmd.Start()
	if err != nil {
		return false
	}

	go func() {
		time.Sleep(time.Second)

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			message := check_progress(job, line)
			printf("\033[2K\r%s", message)

			program_state := check_errors(line)
			if program_state != ALL_GOOD {
				fmt.Println("error", program_state)
				break
			}
		}
	}()

	err = cmd.Wait()

	if err != nil {
		return false
	}

	// close job
	job.Complete = true
	return true
}

const path_rewriter = `
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

output_path = bpy.path.native_pathsep("%s")
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

shortest_common = min(common_paths, key=len)

for node in bpy.context.scene.node_tree.nodes:
    if node.mute or "ignore" in node.label.lower():
        continue
    if node.type == 'OUTPUT_FILE':
        if is_abs(node.base_path):
            continue
        node_path = abs_path(node.base_path)[len(shortest_common) + 1:]
        node.base_path = join(output_path, node_path) + os.sep

bpy.context.scene.render.filepath = join(output_path, source_path[len(shortest_common) + 1:])
`

const (
	py_true  = "True\n"
	py_false = "False\n"
)

func injected_expression(project_dir string, job *Job) string {
	buffer := strings.Builder {}
	buffer.Grow(512)

	buffer.WriteString("import bpy\n")

	if job.Output_Path != "." {
		buffer.WriteString(fmt.Sprintf(path_rewriter, filepath.ToSlash(filepath.Join(project_dir, job.Output_Path))))
	}

	// auto-tiling for Blender 3+
	buffer.WriteString("bpy.context.scene.cycles.use_auto_tile = (bpy.app.version[0] < 3)\n")

	// always remove placeholders because it interferes with restarts
	buffer.WriteString("bpy.context.scene.render.use_placeholder = False\n")

	// @todo experimental for render time testing
	buffer.WriteString("bpy.context.scene.render.use_render_cache = True\n")

	if job.Resolution_X > 0 && job.Resolution_Y > 0 {
		buffer.WriteString(fmt.Sprintf("bpy.context.scene.render.resolution_x = %d\n", job.Resolution_X))
		buffer.WriteString(fmt.Sprintf("bpy.context.scene.render.resolution_y = %d\n", job.Resolution_Y))
		buffer.WriteString("bpy.context.scene.render.resolution_percentage = 100\n")
	}

	// whether to overwrite extant frames
	// @todo currently always false unless manually edited in the order file
	buffer.WriteString("bpy.context.scene.render.use_overwrite = ")
	if job.Overwrite {
		buffer.WriteString(py_true)
	} else {
		buffer.WriteString(py_false)
	}

	return buffer.String()
}

type sous_error uint8
const (
	ALL_GOOD sous_error = iota
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

func (e sous_error) String() string {
	switch e {
	case NO_MEMORY:
		return "Out of RAM"
	case NO_VIDEO_MEMORY:
		return "Out of VRAM"
	case FILESYSTEM_ERROR:
		return "Failed to read data from filesystem"
	case PYTHON_FAIL:
		return "Python failed to initialise"
	case RENDERER_CRASH:
		return "Renderer crashed"
	case RENDERER_NOT_SUPPORTED:
		return "Renderer not supported"
	case RENDERER_KERNEL_FAIL:
		return "CUDA kernel failed to compile"
	case GPU_NOT_SUPPORTED:
		return "Graphics card not supported"
	case EXCEPTION_ACCESS_VIOLATION:
		// this one is non-specific, so passing it on directly
		// will help people find the real answer faster by
		// searching themselves — it could be drivers, display
		// properties and (notably) some weird Windows quirks
		return "EXCEPTION_ACCESS_VIOLATION"
	}
	return "" // never happens
}

func check_errors(input string) sous_error {
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

func check_progress(job *Job, input string) string {
	buffer := strings.Builder {}

	if strings.HasPrefix(input, "Fra:") {
		buffer.Grow(64)
		buffer.WriteString("Frame: ")

		for i, c := range input {
			if unicode.IsSpace(c) {
				the_frame := input[4:i]

				percentage, ok := parse_uint(the_frame)

				if ok {
					percentage = uint(float64(percentage - job.Start_Frame) / float64(job.frame_count) * 100)
				}

				if strings.Contains(input, "Compositing") {
					buffer.WriteString(fmt.Sprintf("%s (Composite) — %d%%", the_frame, percentage))
				} else {
					buffer.WriteString(fmt.Sprintf("%s (Render) — %d%%", the_frame, percentage))
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

				buffer.WriteString(" — ")
				buffer.WriteString(the_time)
			}
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