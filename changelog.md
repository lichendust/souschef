# Changelog

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
