package main

import (
	"os"
	"fmt"
	"time"
	"sort"
	"bytes"
	"path/filepath"

	"io/fs"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type Job struct {
	Name           hash      `toml:"name"`
	Blender_Target hash      `toml:"blender_target"`
	Time           time.Time `toml:"time"`

	Start_Frame uint         `toml:"start_frame"`
	End_Frame   uint         `toml:"end_frame"`
	Frame_Count uint         `toml:"frame_count"`

	Source_Path string       `toml:"source_path"`
	Target_Path string       `toml:"target_path"`
	Output_Path string       `toml:"output_path"`

	Overwrite bool           `toml:"overwrite"`

	// internal
	Complete  bool           `toml:"complete"`
}

type job_array []*Job

func (jobs job_array) Len() int {
	return len(jobs)
}
func (jobs job_array) Less(i, j int) bool {
	return jobs[i].Time.Before(jobs[j].Time)
}
func (jobs job_array) Swap(i, j int) {
	jobs[i], jobs[j] = jobs[j], jobs[i]
}

func (job *Job) String() string {
	return fmt.Sprintf("[%s]\nsource %s\ntarget %s\noutput %s\n", job.Name.word, job.Source_Path, job.Target_Path, job.Output_Path)
}

func serialise_job(job *Job, file_path string) {
	buffer := bytes.Buffer {}
	buffer.Grow(256)

	if err := toml.NewEncoder(&buffer).Encode(job); err != nil {
	    panic(err)
	}

	if err := ioutil.WriteFile(file_path, buffer.Bytes(), 0777); err != nil {
		panic(err)
	}
}

func unserialise_job(path string) (*Job, bool) {
	blob, ok := load_file(path)

	if !ok {
		fmt.Fprintf(os.Stderr, "failed to read job at %q\n", path)
		return nil, false
	}

	data := Job {}

	{
		_, err := toml.Decode(blob, &data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse job at %q\n", path)
			return nil, false
		}
	}

	return &data, true
}

func load_jobs(root string, shallow bool) ([]*Job, bool) {
	job_list := make(job_array, 0, 16)

	root = filepath.Join(root, jobs_dir)

	first := true
	err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		if first {
			first = false
			return nil
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		if shallow {
			job_list = append(job_list, &Job {
				Name: new_hash(info.Name()),
			})
			return nil
		}

		if x, ok := unserialise_job(path); ok {
			job_list = append(job_list, x)
		} else {
			panic(path) // @error
		}

		return nil
	})

	if err != nil {
		// @error
		return nil, false
	}

	sort.Sort(job_list)

	return job_list, true
}