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

# check the reported file class matches what's expected
function check() {
  actual=$(file -b ${1});
  if [[ "${actual}" != "${2}"* ]]; then
    echo "invalid file architecture: ${1}"
    echo "expected: ${2}"
    echo "actual: ${actual}"
    echo "-${actual}-${2}-"
    exit 1
  fi
}

# if the 'UPX' bindary packer is available, use it
# https://upx.github.io/
function compress() {
  [ -x "$(command -v upx)" ] && upx "${1}"
}

# ensure build directory exists
rm -rf build; mkdir -p build

echo "[compile] linux arm64";
env GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -gcflags=-trimpath="${GOPATH}" -asmflags=-trimpath="${GOPATH}";
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-arm64;
check 'build/pbf2json.linux-arm64' 'ELF 64-bit LSB executable';
compress 'build/pbf2json.linux-arm64';

echo "[compile] linux x64";
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -gcflags=-trimpath="${GOPATH}" -asmflags=-trimpath="${GOPATH}";
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-x64;
check 'build/pbf2json.linux-x64' 'ELF 64-bit LSB executable';
compress 'build/pbf2json.linux-x64';

echo "[compile] darwin x64";
env GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -gcflags=-trimpath="${GOPATH}" -asmflags=-trimpath="${GOPATH}";
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.darwin-x64;
check 'build/pbf2json.darwin-x64' 'Mach-O 64-bit';
# UPX disabled due to https://github.com/upx/upx/issues/187
# compress 'build/pbf2json.darwin-x64';

echo "[compile] darwin arm64";
env GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -gcflags=-trimpath="${GOPATH}" -asmflags=-trimpath="${GOPATH}";
assert $?;
chmod +x pbf2json;
mv pbf2json build/pbf2json.darwin-arm64;
check 'build/pbf2json.darwin-arm64' 'Mach-O 64-bit';
# UPX disabled due to https://github.com/upx/upx/issues/187
# note: this is untested, assumed to be the same bug as above
# compress 'build/pbf2json.darwin-arm64';

echo "[compile] windows x64";
env GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -gcflags=-trimpath=${GOPATH} -asmflags=-trimpath=${GOPATH} -o pbf2json.exe;
assert $?;
chmod +x pbf2json.exe;
mv pbf2json.exe build/pbf2json.win32-x64;
check 'build/pbf2json.win32-x64' 'PE32+ executable (console) x86-64';
# UPX disabled for windows
# compress 'build/pbf2json.win32-x64';
