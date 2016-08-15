#!/bin/bash

# note: you will need to follow the guide here to install the tools required
# to cross-compile for different architectures on your host machine.
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5

# ensure the compiler exits with status 0
function assert() {
  if [[ $1 != 0 ]]; then
    echo "one or more architectures failed to compile"
    exit $1;
  fi
}

echo "[compile] linux arm";
env GOOS=linux GOARCH=arm go build;
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-arm;

# echo "[compile] linux i386";
# env GOOS=linux GOARCH=386 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.linux-ia32;

echo "[compile] linux x64";
env GOOS=linux GOARCH=amd64 go build;
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-x64;

# echo "[compile] darwin i386";
# env GOOS=darwin GOARCH=386 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.darwin-ia32;

echo "[compile] darwin x64";
env GOOS=darwin GOARCH=amd64 go build;
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.darwin-x64;

# echo "[compile] windows i386";
# env GOOS=windows GOARCH=386 go build;
# assert $?;
# chmod +x pbf2json.exe;
# mv pbf2json.exe build/pbf2json.win32-ia32;

echo "[compile] windows x64";
env GOOS=windows GOARCH=386 go build -o pbf2json.exe;
assert $?;
chmod +x pbf2json.exe;
mv pbf2json.exe build/pbf2json.win32-x64;

# echo "[compile] freebsd arm";
# env GOOS=freebsd GOARCH=arm go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-arm;

# echo "[compile] freebsd i386";
# env GOOS=freebsd GOARCH=386 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-ia32;

# echo "[compile] freebsd x64";
# env GOOS=freebsd GOARCH=amd64 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-x64;

# echo "[compile] openbsd i386";
# env GOOS=openbsd GOARCH=386 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.openbsd-ia32;

# echo "[compile] openbsd x64";
# env GOOS=openbsd GOARCH=amd64 go build;
# assert $?;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.openbsd-x64;
