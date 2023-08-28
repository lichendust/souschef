# 🍱 Sous Chef

Sous Chef is a rendering assistant for large Blender projects.

It takes care of queuing scenes for rendering, allowing large batches to be paused and resumed, wrangles outputs (especially file nodes!) and generally makes offline rendering simpler for a solo artist or a small team.

## Table of Contents

<!-- MarkdownTOC autolink="true" -->

- [Manifesto](#manifesto)
	- [Sous Chef?](#sous-chef)
- [Usage](#usage)
- [Basic Commands](#basic-commands)
	- [Init](#init)
	- [Order](#order)
		- [Output Paths](#output-paths)
	- [List](#list)
	- [Render](#render)
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

<!-- /MarkdownTOC -->

## Manifesto

Unlike most farm tools that focus on distributing and managing render workloads across many computers, like [Flamenco](https://flamenco.io), Sous Chef is the opposite.

Rather than many machines running one job, Sous Chef looks after one machine running many jobs.

To briefly explain, Sous Chef creates a directory — `.souschef` — in the root of a production's repository, most likely alongside a similar version control directory like `.git` or `.hg`.  This directory stores a running list of jobs that are queued in the order they arrive in (presently).  Each job may optionally hold an entire clone of the target scene and its dependencies, allowing work to progress without fear of changing resources during rendering.

This mode allows scenes and their dependencies to be protected and locked while ongoing changes are made to the rest of the project.

While Sous Chef was conceived for single users, it's easy to imagine it being used on a NAS or similar file share; a small team working from a shared drive or version control system with a single, large render node (like a big desktop with a couple of GPUs in it) that can be requested to perform their queued renders as needed.

If the `.souschef` project is hosted on that same file share, Sous Chef can submit jobs centrally. The NAS itself will ensure all jobs are "published" to all users and the queue order is visible. That one beefy node can then be triggered as needed to work through the queue.

Sous Chef also creates a temporary lock file in each order while it's being processed, with the hostname of the machine that got to that order first. So while a single render node is recommended, this actually enable multiple nodes to safely execute in parallel on the same file share, skipping any locked orders. This isn't a complex load-balanced system, but rather you can imagine this as several artists kicking off their machines to render before going home for the night:

	souschef render; shutdown

In any multi-user scenario, Sous Chef is designed to leverage your existing file share infrastructure. It does not create or pollute your network with additional complexity. Simplicity is king.

### Sous Chef?

*'Sous Chef'* is an oblique reference to an analogy I use to explain the rendering step to people who actually go outside: If modelling and animating are the cooking, then rendering is putting it all in the oven at the end.  If you're the head chef, Sous Chef is the... well, sous.

## Usage

Sous Chef is a single, portable binary that *tries* to contain everything.  It does optionally depend on Python for the caching feature — [see below](#blender-asset-tracer).

## Basic Commands

The base Sous Chef commands, which should always be the first argument, are:

- `init`
- `order`
- `list`
- `render`
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
- **Cache**: It can cache an orders's files using [Blender Asset Tracer](#blender-asset-tracer), eliminating concurrency risks at the cost of disk space (a single order could feasibly require a full clone of the entire project, doubling the required disk space for the lifespan of the order).

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

However, the output path of an order is capable of taking into account file nodes in addition to standard compositor output.  If your scene has file nodes, you should *not* use a fully qualified path and instead only supply a directory for the output:

	some/sequence/

A scene with file nodes should be set up for normal use: the file should work when regular GUI rendering is being used.  When the path is then overridden in a Sous Chef order, the program tries its very best to untangle all of the paths and move everything seamlessly to a new output location, preserving the various outputs' own relativity in that new directory.

Consider a Blender file with a file node (containing two sequences):

	path: //../render/04_01/
	seq1: raw_exr/04_01_####.exr
	seq2: shadow_pass/04_01_####.png

— and a regular, file-level output path:

	//../render/04_01/composite/04_01_####.tif

This file can be rendered in GUI without issue.  Now, if a Sous Chef order was to be created with an entirely distinct output, on say, a NAS:

	R:/prod/04_01

— Sous Chef will adjust everything to provide the same relative structure you had locally:

	R:/prod/04_01/raw_exr/04_01_####.exr
	R:/prod/04_01/shadow_pass/04_01_####.png
	R:/prod/04_01/composite/04_01_####.tif

Sous Chef assumes all of the paths in the project are well-formulated and relative; you usually want all your render data coming out in the same place, but absolute paths work too.

Even so, there's a high chance of bugs with complex combinations of file outputs.  Certain odd combinations of absolute and relative paths or mixed absolute mounting points have not been thoroughly tested, so please be careful with complex outputs.

### List

	souschef list

Show a list of the current jobs, active, complete or otherwise.

### Render

	souschef render

Start rendering the currently registered list of jobs.

Creating a order is not *starting* a order.  Once jobs are created, Sous Chef can be instructed to work through the queue.

This allows resources to be allocated as needed: you might process your entire queue overnight on a particularly powerful machine, or you may want to instruct your team to all set their machines rendering as they leave for the day. Sous Chef's lock files ensure multiple machines can cooperate on a queue. You can read more about that [here](#lock-files).

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

Sous Chef also has a very primitive shortcut table, which is currently hard-coded to a few basics (because they're all the ones I use).  I'm documenting them here for now, but these *will almost certainly change*.

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

As stated in the [Manifesto](#manifesto), this is in service of a narrow use-case where multiple machines are processing the same queue.

The machine that first created a lock file can always reopen its own lock files, because Sous Chef assumes that you'll never be silly enough to run multiple instances of the `render` command on the same machine.

This also ensures that in the event of Sous Chef being killed with `ctrl`+`c` or unexpected shutdown — where the lock file will **not** be deleted — the same machine can just continue where it left off when restarted, because it has automatic approval for that order.

For any scenario where it is an issue, `souschef redo [name]` also clears the lock file, freeing the order up.

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

You can use any name you like for each target and create as many targets as you wish.  When you use the `--target` flag, the name is the value you pass.

Right now, this is the config's only purpose, but in future it may support project-wide templating such as output directories or rendering conventions/settings.

## Version Control

If you use project-wide version control, it is recommended to add exclusion rules for `.souschef/orders`, but *check in* the project configuration file `.souschef/config.toml`.

`.souschef` should also be created in the same location as the root of the VCS, alongside `.hg` or `.git`.

## Blender Asset Tracer

In order to use the cache feature, Sous Chef requires a copy of the [Blender Asset Tracer](https://projects.blender.org/blender/blender-asset-tracer).  BAT provides a small suite of tools for inspecting Blender files and their dependencies, automating the rewriting of those connections and packing up scenes and their dependencies to make them wholly portable (and as small as reasonably possible) for render farms.

Sous Chef should not rely on BAT long term.  In an ideal world, BAT would function as an addon or component of Blender with the same stringent upgrade requirements.  As it stands, BAT can sometimes lag behind Blender versions for months or years until a particularly pragmatic Blender developer comes along to maintain it.

The complexity, planned inconsistency and lack of documentation for the Blender file format — a `.blend` is merely a direct serialisation of Blender's entire runtime scene data structure — makes writing an external program that parses it difficult.

This is why Sous Chef actually calls to Blender itself to get information about new orders, by loading the file, having it print relevant information and then closing.  Even Blender's own DNA inspectors [use Blender itself](https://projects.blender.org/blender/blender/src/branch/main/doc/blender_file_format) to write the dump.

Directly porting BAT could be an option, but at 8.5K lines of Python, it's very much a non-trivial exercise that would require the attention of a developer who wholly understands BAT itself and the intricacies of Blender's innards.

### Installing BAT

To be clear, only the cache feature of Sous Chef requires BAT.  For the intended audience of Sous Chef — single artists — it's more than likely that BAT is not necessary.

BAT requires Python 3.10+ (though it seems Python 3+ is generally fine).

1. Run `pip3 install blender-asset-manager`.
2. Ensure your PATH is correctly set up to allow the new `bat` command to run natively.  `pip` should warn if this is not already the case.
3. Sous Chef should now be able to find the BAT command after a restart of your shell.

### Windows Users and the Subsystem

If you are using Windows with the Subsystem for Linux, you'll still need to use the Windows build of Sous Chef and install Windows Python for BAT.  Mixing a Windows copy of Blender with WSL Python and Sous Chef *can work*, but the spaghetti of path mixing is untenable as a maintainer and infuriating to set up correctly for a user.

I strongly recommend against it and will not aid you in supporting it, but it is technically possible. (Hint: Linux `souschef` and Python + BAT, with Windows Blender in the project configuration. Good luck!)

[^1]: The very first version of Sous Chef was born out of the fact that I had a project stuck on proxy rigs in 2.93 but wanted to take advantage of Cycles X during the 3.0 transition.  I could work in 2.93 and render in 3.0 without worrying about accidentally breaking files or opening them in the wrong version and clattering the rigs.  This was during that 3.0-3.2 phase where proxy conversions just made everything worse.  It's less relevant now, but still a useful feature.