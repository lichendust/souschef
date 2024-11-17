# Changelog

## 0.2.0

Full Release

### Features

- Improved the quality of Sous Chef's output: it's now much nicer to look at and provides more information.
- Added configurable placeholder behaviour.

### Bugs

- Overwriting now matches the new placeholder behaviour and correctly defers to the file instead of being always forced in a binary fashion.
- Percentage of output resolution is now accounted for and correctly reported and applied to output.

### Internal

- Removed `targets` sub-command to add new targets from the CLI. You still need to edit the `.souschef/config.toml` yourself anyway, so let's just commit to it.

## 0.1.2

Patch Release

### Features

- Added operating system-specific overrides for configuration, to be used in 'multiple client OSes accessing a single NAS' scenarios.

> This technically breaks semantic versioning, but it's also not a breaking change for any existing user, so I'm doing it anyway.

### Bugs

- Fixed orders with custom outputs but *without file nodes*  getting sent to Siberia on the user's filesystem.  One of those 'this definitely worked when I last looked at it' bugs.

### Internal

- Clarified some text.

## 0.1.1

First Public Release
