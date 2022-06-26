package main

import (
	"fmt"
	"bufio"
	"io/fs"
	"strings"
	"unicode"
	"os/exec"
	"path/filepath"
)

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

func injected_expression(job *Job) string {
	const (
		py_true  = "True\n"
		py_false = "False\n"

		base_bpy  = "import bpy\n"
		tiling    = "bpy.context.scene.cycles.use_auto_tile = "
		overwrite = "bpy.context.scene.render.use_overwrite = "
	)

	buffer := strings.Builder {}
	buffer.Grow(512)

	buffer.WriteString(base_bpy)

	if job.Blender_Target[1] >= '3' {
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

func job_info(job *Job) {
	const expression = `import bpy; print("sous_range", bpy.context.scene.frame_start, bpy.context.scene.frame_end)`

	cmd := exec.Command("C:/Program Files/Blender Foundation/Blender 3.1/blender.exe", "-b", "--python-expr", expression, job.Source_Path)

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

	err = cmd.Wait()

	if err != nil {
		panic(err)
	}
}

func enumerate_installed_versions() []string {
	first := true
	array := make([]string, 0, 8)

	err := filepath.WalkDir(system_path, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		if first {
			first = false
			return nil
		}

		if info.IsDir() {
			array = append(array, info.Name()[8:])
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return array
}