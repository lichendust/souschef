/*
	Sous Chef
	Copyright (C) 2022-2023 Harley Denham

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

import "os"
import "io"
import "fmt"
import "time"
import "sort"
import "bytes"
import "bufio"
import "io/fs"
import "strings"
import "os/exec"
import "math/rand"
import "path/filepath"
import "github.com/BurntSushi/toml"

type Order struct {
	lock string

	Name           string    `toml:"name"`
	Blender_Target string    `toml:"blender_target"`
	Time           time.Time `toml:"time"`

	Start_Frame uint         `toml:"start_frame"`
	End_Frame   uint         `toml:"end_frame"`
	frame_count uint

	Resolution_X uint        `toml:"resolution_y"`
	Resolution_Y uint        `toml:"resolution_x"`

	Source_Path string       `toml:"source_path"`
	Target_Path string       `toml:"target_path"`
	Output_Path string       `toml:"output_path"`

	Overwrite        uint8   `toml:"overwrite"`
	Use_Placeholders uint8   `toml:"use_placeholders"`

	Complete  bool           `toml:"complete"`
}

const (
	UNSPECIFIED uint8 = iota
	YES
	NO
)

const SET_BY_FILE = "[set by file]"

func format_fallback_bool(value uint8) string {
	switch value {
	case YES:
		return "true"
	case NO:
		return "false"
	}
	return SET_BY_FILE
}

func command_order(config *Config, args *Arguments) {
	if !file_exists(args.source_path) {
		eprintf(apply_color("$1%q$0 does not exist.\n"), args.source_path)
		return
	}

	if filepath.Ext(args.source_path) != ".blend" {
		eprintf(apply_color("$1%q$0 is not a Blender file.\n"), args.source_path)
		return
	}

	args.source_path, _ = filepath.Abs(args.source_path)
	args.output_path, _ = filepath.Abs(args.output_path)

	name := args.replace_id
	if name == "" {
		name = new_name(config.project_dir)
	}

	the_order := new(Order)

	the_order.Name = name
	the_order.Time = time.Now()

	the_order.Source_Path = args.source_path
	the_order.Output_Path = args.output_path

	basename := filepath.Base(the_order.Source_Path)

	the_order.Overwrite        = args.overwrite
	the_order.Use_Placeholders = args.use_placeholders

	if args.blender_target == "" {
		if config.Default_Target == "" {
			eprintln("No Blender target has been provided!")
			return
		}
		the_order.Blender_Target = config.Default_Target
	} else {
		the_order.Blender_Target = args.blender_target
	}

	printf("Gathering information from %s...", basename)

	success := order_info(config, the_order)
	if !success {
		eprintf("Failed to gather information from %s!\n", basename)
		return
	}

	printf(RESET_LINE)

	if args.start_frame != 0 && args.end_frame != 0 {
		the_order.Start_Frame = args.start_frame
		the_order.End_Frame   = args.end_frame
	}

	the_order.frame_count = the_order.End_Frame - the_order.Start_Frame

	if args.resolution_x > 0 && args.resolution_y > 0 {
		the_order.Resolution_X = args.resolution_x
		the_order.Resolution_Y = args.resolution_y
	}

	save_path := order_path(config.project_dir, the_order.Name)

	if args.bank_order {
		if !args.is_bat_installed {
			eprintln("BAT is not discoverable on path: the --cache flag will not work.")
			return
		}

		printf("Generating cached copy of %s with BAT...", basename)

		cmd := exec.Command("bat", "pack", the_order.Source_Path, save_path)

		err := cmd.Start()
		if err != nil {
			eprintln("Failed to start BAT!")
			return
		}

		err = cmd.Wait()
		if err != nil {
			eprintln("Failed to cache order using BAT!")
			return
		}

		printf(RESET_LINE)

		the_order.Target_Path = filepath.Join(ORDER_DIR, the_order.Name, basename)
	}

	the_order.Source_Path, _ = filepath.Rel(config.project_dir, the_order.Source_Path)
	the_order.Output_Path, _ = filepath.Rel(config.project_dir, the_order.Output_Path)

	the_order.Source_Path = filepath.ToSlash(the_order.Source_Path)
	the_order.Output_Path = filepath.ToSlash(the_order.Output_Path)

	if !args.bank_order {
		the_order.Target_Path = the_order.Source_Path
		make_directory(save_path)
	}

	save_order(the_order, manifest_path(config.project_dir, the_order.Name))
	printf(apply_color("[$1%s$0] %s"), the_order.Name, basename)

	if args.bank_order {
		if size, ok := dir_size(save_path); ok {
			printf(" | %.2fMB cache size", size)
		}
	}
	printf("\n")
}

// there should be better way to do this, but
// reading Blender files reliably sucks
func order_info(config *Config, order *Order) bool {
	const expression = `import bpy
s = bpy.context.scene
print("sous_range", s.frame_start, s.frame_end)
print("sous_res", s.render.resolution_x, s.render.resolution_y, s.render.resolution_percentage)`

	blender_path, ok := get_blender_path(config, order.Blender_Target)
	if !ok {
		return false
	}

	cmd := exec.Command(blender_path, "-b", order.Source_Path, "--python-expr", expression)

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
				order.Start_Frame = x
			}
			if x, ok := parse_uint(part[1]); ok {
				order.End_Frame = x
			}
		}

		if strings.HasPrefix(line, "sous_res") {
			line = strings.TrimSpace(line[8:])

			part := strings.SplitN(line, " ", 3)

			if x, ok := parse_uint(part[0]); ok {
				order.Resolution_X = x
			}
			if y, ok := parse_uint(part[1]); ok {
				order.Resolution_Y = y
			}

			percentage, ok := parse_uint(part[2])
			if ok {
				m := float64(percentage) / 100
				order.Resolution_X = uint(float64(order.Resolution_X) * m)
				order.Resolution_Y = uint(float64(order.Resolution_Y) * m)
			}

			// @error needed here if we can't parse.
			// I've tried hundreds of files and it's
			// not happened yet so I feel pretty safe
			// in being nasty and leaving it.
		}
	}

	cmd.Wait()
	return true
}

type Order_Array []*Order

func (orders Order_Array) Len() int {
	return len(orders)
}
func (orders Order_Array) Less(i, j int) bool {
	return orders[i].Time.Before(orders[j].Time)
}
func (orders Order_Array) Swap(i, j int) {
	orders[i], orders[j] = orders[j], orders[i]
}

/*func (order *Order) String() string {
	return fmt.Sprintf("[%s]\nsource %s\ntarget %s\noutput %s\n", order.Name, order.Source_Path, order.Target_Path, order.Output_Path)
}*/

func save_order(order *Order, file_path string) bool {
	buffer := bytes.Buffer{}
	buffer.Grow(512)

	if err := toml.NewEncoder(&buffer).Encode(order); err != nil {
		eprintln("Failed to encode order file")
		return false
	}

	if err := os.WriteFile(file_path, buffer.Bytes(), 0777); err != nil {
		fmt.Println(err)
		eprintln("Failed to write order file")
		return false
	}

	return true
}

func load_order(path string) (*Order, bool) {
	blob, ok := load_file(path)
	if !ok {
		return nil, false
	}

	data := new(Order)

	_, err := toml.Decode(blob, data)
	if err != nil {
		panic(err)
		return nil, false
	}

	data.frame_count = data.End_Frame - data.Start_Frame

	return data, true
}

func load_orders(root string, shallow bool) ([]*Order, bool) {
	order_list := make(Order_Array, 0, 16)

	root = filepath.Join(root, ORDER_DIR)

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
				order_list = append(order_list, &Order{
					Name: name,
				})
				return nil
			}

			path = filepath.Join(path, MANIFEST_NAME)

			the_order, ok := load_order(path)
			if !ok {
				// yeah we just have to return _some_ error.
				// unfortunately, Go doesn't provide us with
				// an idiomatic one from the filepath package
				return io.EOF
			}

			blob, ok := load_file(filepath.Join(path, LOCK_NAME))
			if ok {
				the_order.lock = strings.TrimSpace(blob)
			}

			order_list = append(order_list, the_order)

			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		printf("Sous Chef failed to read orders from .souschef!\n")
		return nil, false
	}

	sort.Sort(order_list)

	return order_list, true
}

const NAMES = `ableacidagedalsoareaarmyawaybabybackballbandbankbasebathbearbeatbeenbeerbellbeltbestbillbirdblowblueboatbodybombbondbonebookboombornbossbothbowlbulkburnbushbusycallcalmcamecampcardcarecasecashcastcellchatchipcityclubcoalcoatcodecoldcomecookcoolcopecopycorecostcrewcropdarkdatadatedawndaysdeaddealdeandeardebtdeepdenydeskdialdietdoesdonedoordosedowndrawdrewdropdualdukedustdutyeachearneaseeasteasyedgeelseeveneverevilexitfacefactfailfairfallfarmfastfatefearfeedfeetfellfilefilmfindfinefirefirmfishfiveflatflowfoodfootfordfourfreefromfuelfullfundgaingamegategavegeargenegiftgirlgivegladgoalgoesgoldgolfgonegoodgrewgreygrowgulfhairhalfhallhandhardharmheadhearheatheldhellhereherohighhillhireholeholyhomehopehosthourhugehunthurtideainchintoironitemjackjanejohnjumpjuryjustkeenkeepkentkeptkickkindkingkneeknewknowlackladylaidlakelandlanelastlateleadleftlesslifeliftlikelinkliveloadlocklogolonglooklostloveluckmademailmainmakemanymarkmassmattmealmeanmeetmeremilkmillmindmissmodemoodmoonmostmovemuchmustnamenavynearneednewsnextnicenickninenonenosenoteokayonceonlyontoopenoverpacepackpagepaidpainpairpalmparkpartpastpathpeakpickpinkpipeplanplayplotplugpluspollpoolpoorportpostpullpurepushrailrainrankratereadrealrelyrentrestriceringriskroadrockrollroofroomrootroserulerushsafesaidsakesalesaltsamesandsaveseatseedseekselfsellsentseptshipshopshotshowshutsicksidesignsitesizeslipslowsnowsoftsoilsoldsolesomesongsoonsortsoulspotstarstaystopsuchsuitsuretaketalktanktapetaskteamtechtelltendtermtesttextthattheythinthisthustilltimetinytolltonetonytooktooltourtowntreetriptruetuneturntwintypeunituponuservastviceviewvotewaitwakewalkwallwantwardwarmwashwavewayswearwellwentwerewestwhatwhenwhomwidewifewildwillwindwinkwirewisewishwithwolfwoodwordworeworkyardyawnyearyetiyolkyorkyouryurtzerozestzetazinczonezoom`

func new_name(project_dir string) string {
	for {
		o := rand.Intn(len(NAMES) / 4) * 4
		n := NAMES[o:o + 4]

		if file_exists(order_path(project_dir, n)) {
			continue
		}

		return n
	}
}
