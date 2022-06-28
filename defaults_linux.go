package main

const config_file = `# the version to use by default when creating
# a new job
default_target = "2.93"

# these are example versions that may not
# match your system. please add, remove or
# update paths that are relevant to your
# system
[[blender_target]]
name = "2.93"
path = "~/.software/blender_2.93/blender"

[[blender_target]]
name = "3.1"
path = "~/.software/blender_3.1/blender"

[[blender_target]]
name = "canary"
path = "~/development/buildbot/blender`