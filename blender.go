package main

import (
	"fmt"
	"bufio"
	"os/exec"
	"strings"
	"unicode"
)

func check_progress(input string) string {
	buffer := strings.Builder {} // @todo let's put this buffer in the outer loop above

	if strings.HasPrefix(input, "Fra:") {
		buffer.Grow(64)
		buffer.WriteString("Frame: ")

		for i, c := range input {
			if unicode.IsSpace(c) {
				if strings.Contains(input, "Compositing") {
					buffer.WriteString(fmt.Sprintf("%s (Composite)", input[4:i]))
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

func injected_expression(job *Job) string {
	const (
		py_true  = "True\n"
		py_false = "False\n"
	)

	buffer := strings.Builder {}
	buffer.Grow(512)

	buffer.WriteString("import bpy\n")

	// auto-tiling for Blender 3+
	buffer.WriteString("bpy.context.scene.cycles.use_auto_tile = (bpy.app.version[0] < 3)\n")

	// whether to overwrite extant frames
	// (@todo doesn't seem to be working?)
	buffer.WriteString("bpy.context.scene.render.use_overwrite = ")
	if job.Overwrite {
		buffer.WriteString(py_true)
	} else {
		buffer.WriteString(py_false)
	}

	return buffer.String()
}

func job_info(job *Job) {
	const expression = `import bpy; print("sous_range", bpy.context.scene.frame_start, bpy.context.scene.frame_end)`

	cmd := exec.Command("C:/Program Files/Blender Foundation/Blender 2.93/blender.exe", "-b", job.Source_Path, "--python-expr", expression)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "sous_range") {
			line = strings.TrimSpace(line[10:])

			part := strings.SplitN(line, " ", 2)

			if x, ok := parse_uint(part[0]); ok {
				job.Start_Frame = x
			}
			if x, ok := parse_uint(part[1]); ok {
				job.End_Frame = x
			}
		}
	}

	cmd.Wait()
}