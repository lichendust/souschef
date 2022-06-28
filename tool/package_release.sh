#!/bin/bash

build_dir="build"

set -e

if [ -z $1 ]; then
	echo "please specify a version"
	exit 1;
fi

rm -f $build_dir/*.zip
rm -f $build_dir/*.sha512sum

printf "[packaging]\n"

for f in $build_dir/*; do
	base=$(basename $f)

	echo $base

	name=${base/"_"/"_$1_"}

	# cp -n license $f/license.txt

	pushd $f > /dev/null
	zip -r "../$name.zip" * > /dev/null
	popd > /dev/null

	pushd $build_dir > /dev/null
	sha512sum "$name.zip" > "$name.sha512sum"
	popd > /dev/null
done

printf "\n[checksums]\n"

pushd $build_dir > /dev/null
sha512sum -c *.sha512sum
popd > /dev/null