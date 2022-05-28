package main

import (
	"bufio"
	"unsafe"
	"strings"
	"reflect"
)

type job struct {
	job_name       string
	blender_target uint8

	source_path string
	output_path string

	complete   bool
	overwrite  bool // internal only
}

/*func (job *job) String() {
	buffer := strings.Builder {}

	t := reflect.ValueOf(*job)

	buffer.Grow(128 * t.NumField())

	for i := 0; i < t.NumField(); i++ {
	    // ft := t.Field(i)

		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

	    // fmt.Println(ft)

		varName  := t.Type().Field(i).Name
		varType  := t.Type().Field(i).Type
		// varValue := t.Field(i).Interface()

		fmt.Printf("%v %v\n", varName, varType)

		x := t.Field(i).Interface().(string)

		fmt.Println(x)

	    buffer.WriteString(ft.Name)
	    buffer.WriteString(" = ")
	    buffer.WriteString(ft.Interface().String())
	    buffer.WriteRune('\n')
	}

	// fmt.Println(buffer.String())
}*/

/*func save_jobs(jobs []*Job, file string) {
	for _, job := range jobs {
		if job.Complete {
			continue
		}

		path := fmt.Sprintf("%s/%s.toml", jobs_path, job.job_name)
		err  := ioutil.WriteFile(file, []byte(job.String()), 0777)

		if err != nil {
			panic(err)
		}
	}
}*/

func parse_job(blob string) *job {
	file_buffer := strings.NewReader(blob)

	new_job := job {}
	structure := reflect.ValueOf(&new_job).Elem()

	{
		scanner := bufio.NewScanner(file_buffer)

		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			if line == "" {
				continue
			}

			part := strings.SplitN(line, "=", 2)

			if len(part) != 2 {
				panic("bad line")
			}

			k := strings.TrimSpace(part[0])
			v := strings.TrimSpace(part[1])

			field := structure.FieldByName(k)
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem() // unsafe go brr

			field.SetString(v)
		}
	}

	return &new_job
}