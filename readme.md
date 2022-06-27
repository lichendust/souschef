# 🍱 Sous Chef

Sous Chef is a rendering assistant for large Blender projects with small teams.

It helps with partitioning scenes and dependencies for rendering, as well as queueing and managing batches of renders, all with the goal of avoiding "my workstation is tied up right now" problems that sap valuable working time.

## Table of Contents

<!-- MarkdownTOC autolink="true" -->

- [Manifesto](#manifesto)
- [Sous Chef?](#sous-chef)
- [Usage](#usage)
	- [Jobs](#jobs)
	- [Render Queue](#render-queue)
- [Version Control](#version-control)
- [Blender Asset Tracer](#blender-asset-tracer)
	- [Installing BAT](#installing-bat)
	- [Windows Users and the Subsystem](#windows-users-and-the-subsystem)

<!-- /MarkdownTOC -->

## Manifesto

Unlike most farm tools that focus on distributing and managing render workloads across many computers, like [Flamenco](https://flamenco.io), Sous Chef is the opposite.

Rather than many machines running one job, Sous Chef looks after one machine running many jobs.

To briefly explain, Sous Chef creates a directory — `.souschef` — in the root of a production's repository, most likely alongside a similar version control directory like `.git` or `.hg`.  This directory stores a running list of jobs that are queued in the order they arrive in (presently).  Each job *may* hold an entire clone of the target scene and its dependencies, allowing work to progress without fear of changing resources during rendering.

On a personal device, this allows jobs to be "banked" during a day's work, then running the queue overnight.  On a dedicated computer connected to a NAS, Sous Chef can watch the job directory, allowing multiple artists to submit jobs using the command, rendered on a first-come first-served basis, like a build server.

## Sous Chef?

*Sous Chef* is an oblique reference to an analogy I use to explain 'rendering' 3D scenes to people who actually go outside: If modelling and animating are the cooking, then rendering is putting it all in the oven at the end.

## Usage

Sous Chef is a single, portable binary.

### Jobs

Sous Chef, in its current form, can act in one of two ways in regards to job creation:

- **Live copy**: It can create a job in-place, using the working copy of the film on disk, with obvious concurrency risks (editing assets could cause issues with the ongoing job).
- **Cache**: It can cache a job's files using [BAT](https://developer.blender.org/source/blender-asset-tracer/browse/master), eliminating concurrency risks at the cost of disk space (a single job could feasibly require a full clone of the entire project, doubling the required disk space for the lifespan of the job).

### Render Queue

Creating a job is not *starting* a job.  Sous Chef can, once jobs have been created, start them in two ways:

- Start and render the job queue, exiting when finished.
- Start and render the job queue, remaining alive and watching the job directory for new ones to be added by other instances of Sous Chef.

The latter allows Sous Chef to run in "server mode" without requiring any actual server configuration.  Pointing an additional rendering machine at a NAS fileshare directory shared by multiple users allows it to watch for their ongoing jobs.  Each user simply uses Sous Chef as they would locally.  The job is then written onto the fileshare, where the host computer will scoop up the new jobs and add them to the end of the queue.

## Version Control

If you use project-wide version control, it is recommended to add exclusion rules for —

+ `.souschef/jobs`
+ `.souschef/data`

— but check in everything else, like the project configuration file.

## Blender Asset Tracer

In order to use the cache feature, Sous Chef requires a copy of the [Blender Asset Tracer](https://developer.blender.org/source/blender-asset-tracer/browse/master/).  BAT provides a small suite of tools for inspecting Blender files and their dependencies, automating the rewriting of those connections and packing up scenes and their dependencies to make them wholly portable (and as small as reasonably possible) for render farms.

Sous Chef should not rely on BAT long term.  In an ideal world, BAT would function as an addon or component of Blender with the same stringent upgrade requirements.  As it stands, BAT can sometimes lag behind Blender versions for months or years until a particularly pragmatic Blender developer comes along to maintain it.

I've made some passing attempts at porting BAT to Go for direct inclusion into Sous Chef, but the lack of documentation for the Blender file format — a `.blend` is merely a direct serialisation of Blender's entire runtime scene data structure — and the complexity of BAT, which is 8.5K lines of Python and does significantly more than Sous Chef actually requires, makes maintaining it too time-consuming to not merely recommend just installing BAT.

### Installing BAT

To be clear, only the cache feature of Sous Chef requires BAT.  For the intended audience of Sous Chef — single artists — it's more than likely that BAT is not necessary.

BAT requires Python 3.10+ (though it seems Python 3+ is generally fine).

1. Run `pip3 install blender-asset-manager`.
2. Ensure your PATH is correctly set up to allow the new `bat` command to run natively.  `pip` should warn if this is not already the case.
3. Sous Chef should now be able to find the BAT command after a restart of your shell.

### Windows Users and the Subsystem

If you are using Windows with the Subsystem for Linux, you'll still need to use the Windows build of Sous Chef and install Windows Python.  Mixing a Windows copy of Blender with WSL Python and Sous Chef *could work*, but the spaghetti of path mixing is untenable for me as a maintainer and infuriating to set up correctly for a user.

You can still use `souschef.exe` through WSL, as I do, which works perfectly.