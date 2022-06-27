package main

const default_config_file = `# the version to use by default when creating
# a new job
default_target = "2.93"

# these are example versions that may not
# match your system. please add, remove or
# update paths that are relevant to your
# system
[[blender_target]]
name = "2.93"
path = "C:/Program Files/Blender Foundation/Blender 2.93/blender.exe"

[[blender_target]]
name = "3.1"
path = "C:/Program Files/Blender Foundation/Blender 3.1/blender.exe"

[[blender_target]]
name = "canary"
path = "X:/development/buildbot/blender.exe"`