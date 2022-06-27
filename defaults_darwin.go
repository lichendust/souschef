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
path = "/Applications/Blender 2.93.app/Contents/MacOS/blender"

[[blender_target]]
name = "3.1"
path = "/Applications/Blender 3.1.app/Contents/MacOS/blender"

[[blender_target]]
name = "canary"
path = "/Volumes/Development/buildbot/blender.exe"`