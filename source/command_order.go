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
	"fmt"
	"time"
	"os/exec"
	"path/filepath"
)

func command_order(project_dir string, args *arguments) {
	config, ok := load_config(filepath.Join(project_dir, config_path))
	if !ok {
		return
	}

	args.source_path, _ = filepath.Abs(args.source_path)
	args.output_path, _ = filepath.Abs(args.output_path)

	name := args.replace_id
	if name == "" {
		name = new_name(project_dir)
	}

	the_job := &Job{
		Name:        name,
		Time:        time.Now(),
		Source_Path: args.source_path,
		Output_Path: args.output_path,
	}

	printf(apply_color("creating order \"$1%s$0\" for %s\n"), the_job.Name, filepath.Base(args.source_path))

	if args.blender_target == "" {
		if config.Default_Target == "" {
			eprintln("no valid Blender target in config.toml, or specified as an argument")
			return
		}

		the_job.Blender_Target = config.Default_Target
	} else {
		the_job.Blender_Target = args.blender_target
	}

	if args.start_frame == 0 && args.end_frame == 0 {
		fmt.Printf("reading info from file...")
		job_info(the_job)
		fmt.Printf("\033[2K\r")
	} else {
		the_job.Start_Frame = args.start_frame
		the_job.End_Frame   = args.end_frame
	}

	the_job.frame_count = the_job.End_Frame - the_job.Start_Frame

	if args.resolution_x > 0 && args.resolution_y > 0 {
		the_job.Resolution_X = args.resolution_x
		the_job.Resolution_Y = args.resolution_y
	}

	if args.bank_job {
		fmt.Printf("generating cache copy...")

		pack_path := order_path(project_dir, the_job.Name)

		cmd := exec.Command("bat", "pack", the_job.Source_Path, pack_path)

		err := cmd.Start()
		if err != nil {
			panic(err)
		}

		err = cmd.Wait()
		if err != nil {
			panic(err)
		}

		the_job.Target_Path = filepath.Join(order_dir, the_job.Name, filepath.Base(the_job.Source_Path))

		fmt.Printf("\033[2K\r")

		if size, ok := dir_size(pack_path); ok {
			fmt.Printf("cache size is %.2fMB on disk\n", size)
		}
	}

	the_job.Source_Path, _ = filepath.Rel(project_dir, the_job.Source_Path)
	the_job.Output_Path, _ = filepath.Rel(project_dir, the_job.Output_Path)

	the_job.Source_Path = filepath.ToSlash(the_job.Source_Path)
	the_job.Output_Path = filepath.ToSlash(the_job.Output_Path)

	if !args.bank_job {
		the_job.Target_Path = the_job.Source_Path
		make_directory(order_path(project_dir, the_job.Name))
	}

	serialise_job(the_job, manifest_path(project_dir, the_job.Name))

	fmt.Println("finished!")
}