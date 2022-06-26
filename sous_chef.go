package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"path/filepath"
)

type sous_chef struct {
	queue       []*Job
	project_dir string
}

func main() {
	args, ok := get_arguments()

	if !ok {
		return
	}

	sous := &sous_chef {}

	switch args.command {
	case COMMAND_INIT:
		fmt.Println("not implemented yet!")
		return

	case COMMAND_HELP:
		fmt.Println("not implemented yet!")
		return

	case COMMAND_VERSION:
		fmt.Println("not implemented yet!")
		return
	}

	if path, ok := find_project_dir(); ok {
		sous.project_dir = path
	} else {
		fmt.Fprintln(os.Stderr, "not a Sous Chef project!")
		return
	}

	switch args.command {
	case COMMAND_JOB:
		args.source_path = filepath.Join(sous.project_dir, args.source_path)
		args.output_path = filepath.Join(sous.project_dir, args.output_path)

		the_job := &Job {
			Name:        new_name(),
			Time:        time.Now(),
			Source_Path: args.source_path,
			Output_Path: args.output_path,
		}

		if args.start_frame == 0 && args.end_frame == 0 {
			job_info(the_job)
		} else {
			the_job.Start_Frame = args.start_frame
			the_job.End_Frame   = args.end_frame
		}

		the_job.Frame_Count = the_job.End_Frame - the_job.Start_Frame

		if args.bank_job {
			the_job.Target_Path    = filepath.Join(sous.project_dir, data_dir, the_job.Name.word)
			the_job.Target_Path, _ = filepath.Rel(sous.project_dir, the_job.Target_Path)
			the_job.Target_Path    = filepath.Join(the_job.Target_Path, filepath.Base(the_job.Source_Path))
		}

		the_job.Source_Path, _ = filepath.Rel(sous.project_dir, the_job.Source_Path)
		the_job.Output_Path, _ = filepath.Rel(sous.project_dir, the_job.Output_Path)

		serialise_job(the_job, filepath.Join(sous.project_dir, jobs_dir, the_job.Name.word))

		fmt.Printf("created new job %q for scene %q\n", the_job.Name, filepath.Base(the_job.Source_Path))

	case COMMAND_LIST:
		sous.queue = load_jobs(sous.project_dir, false)

		fmt.Println("jobs   target file")
		fmt.Println("----   -----------")
		for _, job := range sous.queue {
			fmt.Println(job.Name, " ", filepath.Base(job.Source_Path), job.Start_Frame, job.End_Frame)
		}

	case COMMAND_REMOVE:
		// delete job file here
	}
}

func (sous *sous_chef) run_job(job *Job) {
	cmd := build_command(job)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(time.Second)

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			message := check_progress(line)
			fmt.Printf("\r%s - %s", job.Name, message)

			program_state := check_errors(line)
			if program_state != ALL_GOOD {
				fmt.Println("ERROR", program_state)
				break
			}
		}
	}()

	err = cmd.Wait()

	if err != nil {
		panic(err)
		return
	}

	// close job
	job.complete = true
}

/*func (sous *sous_chef) serialise_jobs() {
	for _, job := range sous.queue {}
}*/