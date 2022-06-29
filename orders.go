package main

import (
	"os"
	"fmt"
	"time"
	"sort"
	"bytes"
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
	Frame_Count uint         `toml:"frame_count"`

	Source_Path string       `toml:"source_path"`
	Target_Path string       `toml:"target_path"`
	Output_Path string       `toml:"output_path"`

	Overwrite bool           `toml:"overwrite"`

	// internal
	Complete  bool           `toml:"complete"`
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

func (job *Job) String() string {
	return fmt.Sprintf("[%s]\nsource %s\ntarget %s\noutput %s\n", job.Name.word, job.Source_Path, job.Target_Path, job.Output_Path)
}

func serialise_job(job *Job, file_path string) bool {
	buffer := bytes.Buffer {}
	buffer.Grow(512)

	if err := toml.NewEncoder(&buffer).Encode(job); err != nil {
		fmt.Fprintln(os.Stderr, "failed to encode job file")
		return false
	}

	if err := ioutil.WriteFile(file_path, buffer.Bytes(), 0777); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write job file")
		return false
	}

	return true
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

func new_name() hash {
	o := rand.Intn(20) * 4
	return new_hash(names[o:o + 4])
}