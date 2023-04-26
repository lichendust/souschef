package main

import (
	"fmt"
	"time"
	"sort"
	"bytes"
	"bufio"
	"strings"
	"os/exec"
	"math/rand"
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
	frame_count uint

	Resolution_X uint        `toml:"resolution_y"`
	Resolution_Y uint        `toml:"resolution_x"`

	Source_Path string       `toml:"source_path"`
	Target_Path string       `toml:"target_path"`
	Output_Path string       `toml:"output_path"`

	Overwrite bool           `toml:"overwrite,omitempty"`

	// internal
	Complete  bool           `toml:"complete,omitempty"`
}

// there should be better way to do this, but
// reading Blender files reliably sucks
func job_info(job *Job) {
	const expression = `import bpy; print("sous_range", bpy.context.scene.frame_start, bpy.context.scene.frame_end); print("sous_res", bpy.context.scene.render.resolution_x, bpy.context.scene.render.resolution_y)`

	// @todo make this the default installation in config.toml
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

		if strings.HasPrefix(line, "sous_res") {
			line = strings.TrimSpace(line[8:])

			part := strings.SplitN(line, " ", 2)

			if x, ok := parse_uint(part[0]); ok {
				job.Resolution_X = x
			}
			if x, ok := parse_uint(part[1]); ok {
				job.Resolution_Y = x
			}
		}
	}

	cmd.Wait()
}

func print_order(i int, order *Job) {
	complete := ""
	if order.Complete {
		complete = "complete!"
	}

	printf(apply_color("%d \"$1%s$0\" %-20s %d-%d %dx%d %s\n"),
		i, order.Name, filepath.Base(order.Source_Path),
		order.Start_Frame, order.End_Frame,
		order.Resolution_X, order.Resolution_Y,
		complete)
}

type order_array []*Job

func (orders order_array) Len() int {
	return len(orders)
}
func (orders order_array) Less(i, j int) bool {
	return orders[i].Time.Before(orders[j].Time)
}
func (orders order_array) Swap(i, j int) {
	orders[i], orders[j] = orders[j], orders[i]
}

func (order *Job) String() string {
	return fmt.Sprintf("[%s]\nsource %s\ntarget %s\noutput %s\n", order.Name.word, order.Source_Path, order.Target_Path, order.Output_Path)
}

func serialise_job(order *Job, file_path string) bool {
	buffer := bytes.Buffer {}
	buffer.Grow(512)

	if err := toml.NewEncoder(&buffer).Encode(order); err != nil {
		eprintln("failed to encode order file")
		return false
	}

	if err := ioutil.WriteFile(file_path, buffer.Bytes(), 0777); err != nil {
		fmt.Println(err)
		eprintln("failed to write order file")
		return false
	}

	return true
}

func unserialise_job(path string) (*Job, bool) {
	blob, ok := load_file(path)

	if !ok {
		eprintf("failed to read order at %q\n", path)
		return nil, false
	}

	data := Job{}

	{
		_, err := toml.Decode(blob, &data)
		if err != nil {
			eprintf("failed to parse order at %q\n", path)
			return nil, false
		}
	}

	data.frame_count = data.End_Frame - data.Start_Frame

	return &data, true
}

func load_orders(root string, shallow bool) ([]*Job, bool) {
	job_list := make(order_array, 0, 16)

	root = filepath.Join(root, order_dir)

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
			name := info.Name()

			if shallow {
				job_list = append(job_list, &Job {
					Name: new_hash(name),
				})
				return nil
			}

			path = filepath.Join(path, order_name)

			if x, ok := unserialise_job(path); ok {
				job_list = append(job_list, x)
			} else {
				panic(path) // @error
			}

			return filepath.SkipDir
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

const names = `ableacidagedalsoareaarmyawaybabybackballbandbankbasebathbearbeatbeenbeerbellbeltbestbillbirdblowblueboatbodybombbondbonebookboombornbossbothbowlbulkburnbushbusycallcalmcamecampcardcarecasecashcastcellchatchipcityclubcoalcoatcodecoldcomecookcoolcopecopycorecostcrewcropdarkdatadatedawndaysdeaddealdeandeardebtdeepdenydeskdialdietdoesdonedoordosedowndrawdrewdropdualdukedustdutyeachearneaseeasteasyedgeelseeveneverevilexitfacefactfailfairfallfarmfastfatefearfeedfeetfellfilefilmfindfinefirefirmfishfiveflatflowfoodfootfordfourfreefromfuelfullfundgaingamegategavegeargenegiftgirlgivegladgoalgoesgoldgolfgonegoodgrewgreygrowgulfhairhalfhallhandhardharmheadhearheatheldhellhelphereherohighhillhireholeholyhomehopehosthourhugehunthurtideainchintoironitemjackjanejohnjumpjuryjustkeenkeepkentkeptkickkindkingkneeknewknowlackladylaidlakelandlanelastlateleadleftlesslifeliftlikelinklistliveloadlocklogolonglooklostloveluckmademailmainmakemanymarkmassmattmealmeanmeetmeremilkmillmindmissmodemoodmoonmostmovemuchmustnamenavynearneednewsnextnicenickninenonenosenoteokayonceonlyontoopenoverpacepackpagepaidpainpairpalmparkpartpastpathpeakpickpinkpipeplanplayplotplugpluspollpoolpoorportpostpullpurepushrailrainrankratereadrealrelyrentrestriceringriskroadrockrollroofroomrootroserulerushsafesaidsakesalesaltsamesandsaveseatseedseekselfsellsentseptshipshopshotshowshutsicksidesignsitesizeslipslowsnowsoftsoilsoldsolesomesongsoonsortsoulspotstarstaystopsuchsuitsuretaketalktanktapetaskteamtechtelltendtermtesttextthattheythinthisthustilltimetinytolltonetonytooktooltourtowntreetriptruetuneturntwintypeunituponuservastviceviewvotewaitwakewalkwallwantwardwarmwashwavewayswearwellwentwerewestwhatwhenwhomwidewifewildwillwindwinkwirewisewishwithwolfwoodwordworeworkyardyawnyearyetiyolkyorkyouryurtzerozestzetazinczonezoom`

func init() {
	rand.Seed(time.Now().Unix())
}

func new_name(project_dir string) hash {
	o := rand.Intn(20) * 4
	n := names[o:o + 4]

	if file_exists(order_path(project_dir, n)) {
		return new_name(project_dir)
	}

	return new_hash(n)
}