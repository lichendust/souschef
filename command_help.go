package main

const comm_help = `
$1Usage$0
-----

    souschef command [--flags]

$1Commands$0
--------

    $1init$0     initialise new Sous Chef project
    $1new$0      create a new job
    $1list$0     show the list of current jobs
    $1render$0   start the render queue
    $1clean$0    remove finished jobs
    $1help$0     print this message and others

Use $1souschef help [command]$0 for more information on each of 
the above.`

const comm_render = `
Render will start the job queue, rendering each job 
sequentially.  If any job fails, at any stage, it warn the 
failure and progress to the next one in the queue.

$1Usage$0
-----

    souschef $1render$0 [--watch -w]

$1Watch$0
-------------

    $1--watch -w$0

Enabling "watch" mode will cause Sous Chef to remain alive even 
if there are no available jobs, watching the jobs directory for 
new additions.  In this mode Sous Chef *will* delete finished 
jobs and any associated cache data to save space.`