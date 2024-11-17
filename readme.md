# 🍱 Sous Chef

![](https://stuff.lichendust.com/media/souschef.webp)

Sous Chef is a rendering assistant for large Blender projects.

It takes care of queuing scenes for rendering, allowing large batches to be paused and resumed, wrangles outputs (especially File Nodes!) and generally makes offline rendering simpler for a solo artist or a small team.

## Table of Contents

<!-- MarkdownTOC autolink="true" -->

- [Manifesto](#manifesto)
	- [Sous Chef?](#sous-chef)
- [Getting Started](#getting-started)
- [Basic Commands](#basic-commands)
	- [Init](#init)
	- [Order](#order)
		- [Output Paths](#output-paths)
	- [List](#list)
	- [Render](#render)
	- [Redo](#redo)
	- [Delete](#delete)
	- [Clean](#clean)
- [Order Parameters](#order-parameters)
	- [Cache](#cache)
	- [Target](#target)
	- [Replace](#replace)
	- [Resolution](#resolution)
	- [Frame](#frame)
- [Lock Files](#lock-files)
- [Default Configuration](#default-configuration)
- [Version Control](#version-control)
- [Blender Asset Tracer](#blender-asset-tracer)
	- [Installing BAT](#installing-bat)
	- [Windows Users and the Subsystem](#windows-users-and-the-subsystem)
- [Todo](#todo)

<!-- /MarkdownTOC -->

## Manifesto

Unlike most farm tools that focus on distributing and managing render workloads across many computers, like [Flamenco](https://flamenco.io), Sous Chef is the opposite.

Rather than looking after and feeding jobs to many machines running one job, Sous Chef looks after one machine running many jobs in a queue.

To briefly explain, Sous Chef creates a directory — `.souschef` — in the root of a production's repository, most likely alongside a similar version control directory like `.git` or `.hg`.  You then push jobs into the queue, which are tracked in the `.souschef` directory.

### Sous Chef?

*'Sous Chef'* is an oblique reference to an analogy I use to explain the rendering step to people who actually go outside: If modelling and animating are the cooking, then rendering is putting it all in the oven at the end.  If you're the head chef, Sous Chef is the... sous.

## Getting Started

Sous Chef is a single, portable binary that *tries* to contain everything.  It does optionally depend on Python for the caching feature — [see below](#blender-asset-tracer).

1. After installing on your system, navigate to the highest level of your film or VFX project.
2. Run `souschef init` to create a new project.
3. Run `souschef targets` to view the example Blender targets.
4. Edit `.souschef/config.toml` to set up your Blender targets.
5. Navigate through your project and open a new job with `souschef order [file.blend]`.
6. View the current render queue with `souschef list`.
7. Start rendering the whole queue with `souschef render`.

See below for more detail on every command, or use `souschef help [command]` for the built-in instruction manual.

## Basic Commands

The base Sous Chef commands, which should always be the first argument, are:

- `init`
- `order`
- `list`
- `render`
- `redo`
- `delete`
- `clean`

There's also the usual self-explanatory stuff:

- `help`
- `version`

It should be noted that Sous Chef's help command is extremely powerful and provides detailed information about every command, flag and setting available within Sous Chef.

### Init

	souschef init

Initialise a new Sous Chef directory.  This should be done at the top-level of a production's repository.

### Order

A render job in Sous Chef is called an "order". Sous Chef can act in one of two ways in regards to order creation:

- **Live copy**: It can create an order in-place, using the working copy of the film on disk, with obvious concurrency risks (editing assets could cause issues with the ongoing order).
- **Cache**: It can cache an orders's files using [Blender Asset Tracer](#blender-asset-tracer), eliminating concurrency risks at the cost of disk space (a single order could feasibly require a full clone of the entire project, depending on how complex the dependency tree is, thus doubling the required disk space for the lifespan of the order).

You can perform the latter with the `--cache` flag, but that's getting ahead of ourselves.

For now, you can create a new order with:

	souschef order path/to/file.blend

`order` is actually the default expression of Sous Chef, which means you can omit the keyword:

	souschef path/to/file.blend

You can also specify the output location with a second unflagged argument:

	souschef path/to/file.blend some/render/path

#### Output Paths

When specifying an output in a Sous Chef order, you should use a fully qualified Blender output path:

	some/sequence/frame_####.jpg

However, if you are using File Nodes in the Compositor, Sous Chef will account for these.  If your scene has File Nodes, you should *not* use a fully qualified path and instead only supply a directory for the output:

	some/sequence/

Sous Chef will patch up all the various file paths on your system and construct the shortest shared output location. Consider a Blender file with a File Node in the Compositor (containing two sequences):

	path: //../render/04_01/
	seq1: raw_exr/04_01_####.exr
	seq2: shadow_pass/04_01_####.png

— and the standard output panel in the Properties tab:

	//../render/04_01/composite/04_01_####.tif

Now, if a Sous Chef order was to be created with an entirely distinct output, on say, a NAS:

	R:/prod/04_01

— Sous Chef will adjust everything to provide the same relative structure you had locally:

	R:/prod/04_01/raw_exr/04_01_####.exr
	R:/prod/04_01/shadow_pass/04_01_####.png
	R:/prod/04_01/composite/04_01_####.tif

Sous Chef assumes all of the paths in the project are well-formulated and relative; you usually want all your render data coming out in the same place, but absolute paths can work too, so long as they're on the same drive — Sous Chef needs an obvious shared root to fix up the paths.

As with any early software, there's a high chance of bugs with complex combinations of file outputs.  Certain odd combinations of absolute and relative paths have not been thoroughly tested, so please be careful with complex outputs.

### List

	souschef list

Show a list of the current jobs, active, complete or otherwise.

### Render

	souschef render

Start rendering the current queue of orders.

Creating a order is not *starting* a order.  Once jobs are created, Sous Chef can be instructed to work through the queue.

This allows resources to be allocated as needed: you might process your entire queue overnight on a particularly powerful machine.

### Redo

	souschef redo [name]

Resets the 'completed' status of a selected order, moving it to the end of the queue. This allows it to be restarted without needing to fetch or regenerate any new data. Useful if something minor went wrong that can be quickly fixed in place (like a faulty output path).

### Delete

	souschef delete [name]

Instantly deletes the specified order from the queue. It's gone.

### Clean

You can purge the order directory with:

	souschef clean

This only removes finished jobs by default, but you can purge all jobs with the additional `--hard` flag.

## Order Parameters

When creating an order, there are a number of additional options available.

Any values that affect settings which are also defined in a scene file are optional: the file's original settings will be used in their absence.

### Cache

	--cache
	-c

Specifies that this order should be cached, which means packed up into a discreet copy and filed away to protect it from ongoing changes.  This feature requires [BAT](#blender-asset-tracer).

### Target

	--target [name]
	-t [name]

Allows you to choose a Blender target for your order, in case of split versions or compatibility issues on a project[^1].  See the [configuration below](#default-configuration) for the exact meaning of a target.

### Replace

	--replace [name]

Create a new order with new parameters, but specifically overwrite an existing order.  This will *not* keep its timestamp and will bump it to the back of the queue.

Useful for any missed configuration, but it will rebuild the entire order.

### Resolution

	--resolution 1000x1000
	-r 1000x1000

Override the output resolution.  Both X and Y dimensions must be supplied.

Sous Chef also has a very primitive shortcut table, which is currently hard-coded to a few basics that I use myself because I'm very lazy.  I'm documenting them here for now, but these *will change in future* because they need to.

	-r dcp4k

- `UHD` — 3840 x 2160
- `HD` — 1920 x 1080

- `DCP4K` — 4096 x 1716
- `DCP2K` — 2048 x 858

### Frame

	--frame 1:250
	-f 300

Override the frame-range.  If only one value is supplied, it's used as the end frame, with the starting frame assumed to be 1.

## Lock Files

Whenever Sous Chef is actively rendering an order, a `lock.txt` file is created in the order's directory. This lock file contains the hostname of the machine currently hosting the instance of Blender with the file open.

This is in service of a narrow use-case where multiple machines can simultaneously process the same queue, such as on a NAS.

The machine that first created a lock file can always reopen its own lock files, because Sous Chef assumes that you'll never be silly enough to run multiple instances of the `render` command on the same machine.

This also ensures that in the event of Sous Chef being killed with `ctrl`+`c` or unexpected shutdown — where the lock file will **not** be deleted — the same machine can just continue where it left off when restarted, because it has automatic approval for that order.

For any other scenario where this is an issue, `souschef redo <order>` will clear the lock file, freeing the order up.

## Default Configuration

When calling `souschef init`, the default project configuration will look something similar to this, adjusted for your operating system:

```toml
default_target = "2.93"

[[target]]
name = "2.93"
path = "C:/Program Files/Blender Foundation/Blender 2.93/blender.exe"

[[target]]
name = "3.6"
path = "C:/Program Files/Blender Foundation/Blender 3.6/blender.exe"

[[target]]
name = "canary"
path = "X:/dev/buildbot/custom-blender.exe"
```

This configuration is primarily aimed at sorting out Blender versions, especially if you're extremely sensible and lock versions on projects or even distribute internal portable builds to ensure things don't break across artists' computers.

You can use any label — `name` — you like for each target and create as many targets as you wish.  When you use the `--target` flag, the label is the value you pass.

You can also create multiple, operating system level configuration files —

- `config.toml`
- `config_linux.toml`
- `config_macos.toml`
- `config_windows.toml`

`config.toml` is the standard fallback for all operating systems; if you're a single user, or your shop is entirely one operating system, you can just use that.  But if multiple OSes are accessing and rendering for a project on a single shared-volume, you'll want to set their paths accordingly.

Obviously, if you're assuming every order can be fulfilled by any other machine accessing the production, you'll need to make sure your target labels match across all platforms *and* that your installation paths are the same on each machine running the same operating system.

> Right now, all of this is the config's only purpose, but in future it may support project-wide templating such as output directories or rendering conventions/settings.
>
> I'm quite partial to the idea of automatically generated render directories that are templated from the scene file paths.

## Version Control

If you use project-wide version control, it is recommended to add exclusion rules for `.souschef/orders`, but *check in* the configuration `.toml` files.

`.souschef` should also be created in the same location as the root of the VCS, alongside `.hg` or `.git`.  Sous Chef uses the same 'search upwards' mechanism as most VCSes, so you can invoke it from anywhere inside the project hierarchy.

## Blender Asset Tracer

In order to use the cache feature, Sous Chef requires a copy of the [Blender Asset Tracer](https://projects.blender.org/blender/blender-asset-tracer).  BAT provides a small suite of tools for inspecting Blender files and their dependencies, automating the rewriting of those connections and packing up scenes and their dependencies to make them wholly portable (and as small as reasonably possible) for render farms.

Sous Chef should not rely on BAT long term.  In an ideal world, BAT would function as an addon or component of Blender with the same stringent upgrade requirements.  As it stands, BAT can sometimes lag behind Blender versions for months or years until a particularly pragmatic Blender developer comes along to maintain it.

The complexity, planned inconsistency and lack of documentation for the Blender file format — a `.blend` is merely a direct serialisation of Blender's entire runtime scene data structure — makes writing an external program that parses it difficult.

This is why Sous Chef actually calls to Blender itself to get information about new orders, by loading the file, then asking it to print out relevant information.  Even Blender's own DNA inspectors [use Blender itself](https://projects.blender.org/blender/blender/src/branch/main/doc/blender_file_format) to write the dump.

Directly porting BAT could be an option, but at ~8.5K lines of Python, it's very much a non-trivial exercise that would require the attention of a developer who wholly understands BAT itself and the intricacies of Blender's innards.

### Installing BAT

To be clear, only the cache feature of Sous Chef requires BAT.

BAT requires Python 3.10+ (though it seems Python 3+ is generally fine).

1. Run `pip3 install blender_asset_tracer`.
2. Ensure your PATH is correctly set up to allow the new `bat` command to run natively.  `pip` should warn if this is not already the case.
3. Sous Chef should now be able to find the BAT command after a restart of your shell.

### Windows Users and the Subsystem

If you are using Windows with the Subsystem for Linux, you'll still need to use the Windows build of Sous Chef and install Windows Python for BAT.  Mixing a Windows copy of Blender with WSL Python and Sous Chef can work, but the spaghetti of path mixing is untenable as a maintainer and infuriating to set up correctly for a user.

I strongly recommend against it and will not aid you in supporting it, but it is technically possible. (Hint: Linux `souschef` and Python + BAT, with `/mnt/c/` Windows Blender paths in the project configuration. Good luck!)

## Todo

- Scene defaults in project-level config. A small block of resolution/format/frame-ranges/render-levels, etc. that the project manager can label: "previz", "low_quality" "final_dcp", etc.
- The original creator's hostname should appear in orders.

[^1]: The very first version of Sous Chef was born out of the fact that I had a project stuck on proxy rigs in 2.93 but wanted to take advantage of Cycles X during the 3.0 transition.  I could work in 2.93 and render in 3.0 without worrying about accidentally breaking files or opening them in the wrong version and clattering the rigs.  This was during that 3.0-3.2 phase where proxy conversions just made everything worse.  It's less relevant now, but still a useful feature.
