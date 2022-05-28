package main

import (
	"fmt"
	"time"
	"bufio"
	"os/exec"
)

type sous_chef struct {
	backend backend
	queue   []*job
}

type backend interface {
	build_command(*job) *exec.Cmd
	check_progress(string) string
	check_errors(string) sous_error
}

func main() {
	sous := &sous_chef {
		backend_blender {},
		[]*job {
			{
				job_name:    "jeff",
				source_path: "project/scene.blend",
			},
		},
	}

	for _, job := range sous.queue {
		sous.run_job(job)
	}
}

func (sous *sous_chef) run_job(job *job) {
	error_channel := make(chan error, 1)

	cmd := sous.backend.build_command(job)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		error_channel <- cmd.Wait()
	}()

	go func() {
		time.Sleep(time.Second)

		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := scanner.Text()

			{
				message := sous.backend.check_progress(line)
				fmt.Printf(" \r%s - %s", job.job_name, message)
			}

			{
				message := sous.backend.check_errors(line)
				if message != ALL_GOOD {
					fmt.Printf("\n\n\n\n%s", message)
				}
			}
		}
	}()

	select {
	case err := <-error_channel:
		if err != nil {
			panic(err)
		}
	}
}