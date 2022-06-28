package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"os/exec"
	"strings"
	"path/filepath"
)

func main() {
	args, ok := get_arguments()

	if !ok {
		return
	}

	switch args.command {
	case COMMAND_INIT:
		if !make_directory(jobs_dir) {
			return
		}
		if !make_directory(data_dir) {
			return
		}
		if !write_file(config_path, config_file) {
			return
		}
		fmt.Println("initialised Sous Chef project")
		return

	case COMMAND_HELP:
		fmt.Println("not implemented yet!")
		return

	case COMMAND_VERSION:
		fmt.Println("not implemented yet!")
		return

	case COMMAND_REMOVE:
		fmt.Println("not implemented yet!")
		return
	}

	project_dir, ok := find_project_dir()

	if !ok {
		fmt.Fprintln(os.Stderr, "not a Sous Chef project!")
		return
	}

	switch args.command {
	case COMMAND_LIST:
		queue, ok := load_jobs(project_dir, false)

		if !ok {
			return
		}

		if len(queue) == 0 {
			fmt.Println("no jobs found!")
			return
		}
		for i, job := range queue {
			fmt.Println(i + 1, job.Name, " ", filepath.Base(job.Source_Path), " ", job.Start_Frame, job.End_Frame)
		}
		return

	case COMMAND_JOB:
		config, ok := load_config(filepath.Join(project_dir, config_file))

		if !ok {
			return
		}

		args.source_path, _ = filepath.Abs(args.source_path)
		args.output_path, _ = filepath.Abs(args.output_path)

		the_job := &Job {
			Name:        new_name(),
			Time:        time.Now(),
			Source_Path: args.source_path,
			Output_Path: args.output_path,
		}

		if args.blender_target == "" {
			if config.Default_Target.uint32 == 0 {
				fmt.Fprintln(os.Stderr, "no valid Blender target in config.toml, or specified as an argument")
				return
			}

			the_job.Blender_Target = config.Default_Target
		} else {
			the_job.Blender_Target = new_hash(args.blender_target)
		}

		if args.start_frame == 0 && args.end_frame == 0 {
			job_info(the_job)
		} else {
			the_job.Start_Frame = args.start_frame
			the_job.End_Frame   = args.end_frame
		}
		the_job.Frame_Count = the_job.End_Frame - the_job.Start_Frame

		if args.bank_job {
			pack_path := filepath.Join(project_dir, data_dir, the_job.Name.word)

			cmd := exec.Command("bat", "pack", the_job.Source_Path, pack_path)

			err := cmd.Start()
			if err != nil {
				panic(err)
			}

			err = cmd.Wait()
			if err != nil {
				panic(err)
			}

			the_job.Target_Path = filepath.Join(data_dir, the_job.Name.word, filepath.Base(the_job.Source_Path))
		}

		the_job.Source_Path, _ = filepath.Rel(project_dir, the_job.Source_Path)
		the_job.Output_Path, _ = filepath.Rel(project_dir, the_job.Output_Path)

		if !args.bank_job {
			the_job.Target_Path = the_job.Source_Path
		}

		serialise_job(the_job, filepath.Join(project_dir, jobs_dir, the_job.Name.word))

		fmt.Printf("created new job %q for scene %q\n", the_job.Name, filepath.Base(the_job.Source_Path))

		return

	case COMMAND_RENDER:
		config, ok := load_config(filepath.Join(project_dir, config_path))

		if !ok {
			return
		}

		queue, ok := load_jobs(project_dir, false)

		if !ok {
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
				ok := serialise_job(the_job, filepath.Join(project_dir, jobs_dir, the_job.Name.word))

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
	output := filepath.Join(project_dir, job.Output_Path)

	// @todo we don't use the correct outputs yet!!
	cmd := exec.Command(blender_path, "-b", target, "-o", output, "--python-expr", injected_expression(job), "-a")

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

			message := check_progress(line)
			fmt.Printf("\033[2K\r%s", message)

			program_state := check_errors(line)
			if program_state != ALL_GOOD {
				fmt.Println("ERROR", program_state)
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