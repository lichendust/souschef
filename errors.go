package main

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
		return "EXCEPTION_ACCESS_VIOLATION"
		// this one is non-specific, so passing it on directly
		// will help people find the real answer faster by
		// searching themselves — it could be drivers, display
		// properties and (notably) some weird Windows quirks
	}
	return "All Good!"
}