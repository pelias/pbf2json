
#!/bin/bash
# @ref: https://gist.github.com/missinglink/352f5be805395babada0

dirname=$(dirname $0);
cd "$dirname";
file=vancouver_canada.osm.pbf;

if [ -e $file ]; then
    hash=`shasum "$file" | awk '{ print $1 }'`;
    if test "$hash" = c033bef77dcb88ceb8e224aa17c6fe388a217c98; then
        exit 0; # already exists with correct hash
    fi
fi

curl -s -o $file http://peter.johnson.s3.amazonaws.com/vancouver_canada.osm.pbf \
|| (echo "failed to fetch fixture file: $file" >&2; exit 1)