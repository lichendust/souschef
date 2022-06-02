package main

import (
	"fmt"
	"strings"
	"unicode"
	"os/exec"
	// "path/filepath"
)

func bank_command(job *Job) *exec.Cmd {
	return exec.Command("bat", "pack", job.Source_Path, job.Target_Path)
}

func build_command(job *Job) *exec.Cmd {
	return exec.Command("C:/Program Files/Blender Foundation/Blender 3.1/blender.exe", "-b", "--python-expr", build_python_expression(job), job.Source_Path, "-a")
}

func check_progress(input string) string {
	buffer := strings.Builder {}

	if strings.HasPrefix(input, "Fra:") {
		buffer.Grow(64)
		buffer.WriteString("Frame: ")

		for i, c := range input {
			if unicode.IsSpace(c) {
				if strings.Contains(input, "Compositing") {
					buffer.WriteString(fmt.Sprintf("%s (Compositing)", input[4:i]))
				} else {
					buffer.WriteString(fmt.Sprintf("%s (Render)", input[4:i]))
				}
				break
			}
		}
	}

	return buffer.String()
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

func build_python_expression(job *Job) string {
	const (
		py_true  = "True\n"
		py_false = "False\n"

		base      = "import bpy\n"
		tiling    = "bpy.context.scene.cycles.use_auto_tile = "
		overwrite = "bpy.context.scene.render.use_overwrite = "
	)

	buffer := strings.Builder {}
	buffer.Grow(512)

	buffer.WriteString(base)

	if job.Blender_Target >= 3 {
		buffer.WriteString(tiling)
		buffer.WriteString(py_true)
	}

	buffer.WriteString(overwrite)

	if job.Overwrite {
		buffer.WriteString(py_true)
	} else {
		buffer.WriteString(py_false)
	}

	return buffer.String()
}