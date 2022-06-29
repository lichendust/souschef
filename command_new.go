package main

import (
	"os"
	"os/exec"

	"fmt"
	"time"
	"path/filepath"
)

func command_order(project_dir string, args *arguments) {
	config, ok := load_config(filepath.Join(project_dir, config_path))

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

	fmt.Printf("creating job for %s\n", filepath.Base(args.source_path))

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
		fmt.Printf("reading info from file...")
		job_info(the_job)
		fmt.Printf("\033[2K\r")
	} else {
		the_job.Start_Frame = args.start_frame
		the_job.End_Frame   = args.end_frame
	}
	the_job.Frame_Count = the_job.End_Frame - the_job.Start_Frame

	if args.bank_job {
		fmt.Printf("generating cache copy...")

		pack_path := order_path(project_dir, the_job.Name.word)

		cmd := exec.Command("bat", "pack", the_job.Source_Path, pack_path)

		err := cmd.Start()
		if err != nil {
			panic(err)
		}

		err = cmd.Wait()
		if err != nil {
			panic(err)
		}

		the_job.Target_Path = filepath.Join(order_dir, the_job.Name.word, filepath.Base(the_job.Source_Path))

		fmt.Printf("\033[2K\r")

		if size, ok := dir_size(pack_path); ok {
			fmt.Printf("cache size is %.2fMB on disk\n", size)
		}
	}

	the_job.Source_Path, _ = filepath.Rel(project_dir, the_job.Source_Path)
	the_job.Output_Path, _ = filepath.Rel(project_dir, the_job.Output_Path)

	if !args.bank_job {
		the_job.Target_Path = the_job.Source_Path
	}

	serialise_job(the_job, manifest_path(project_dir, the_job.Name.word))

	fmt.Println("finished!")
}