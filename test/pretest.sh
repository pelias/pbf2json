
#!/bin/bash
# @ref: https://gist.github.com/missinglink/414edeaae2298db711e3

dirname=$(dirname `readlink -f $0`);
cd "$dirname";
file=somes.osm.pbf;

if [ -e $file ]; then
    hash=`shasum "$file" | awk '{ print $1 }'`;
    if test "$hash" = f67e1a56ff6b43fefb204df3e1849c5beee9cd24; then
        exit 0; # already exists with correct hash
    fi
fi

curl -s -o $file http://peter.johnson.s3.amazonaws.com/somes.osm.pbf \
|| (echo "failed to fetch fixture file: $file" >&2; exit 1)