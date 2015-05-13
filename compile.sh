#!/bin/bash

# note: you will need to follow the guide here to install the tools required
# to cross-compile for different architectures on your host machine.
# http://dave.cheney.net/2013/07/09/an-introduction-to-cross-compilation-with-go-1-1

# change this path to match your env
source /var/www/golang-crosscompile/crosscompile.bash

echo "[compile] linux arm";
go-linux-arm build pbf2json.go;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-arm;

# echo "[compile] linux i386";
# go-linux-386 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.linux-ia32;

echo "[compile] linux x64";
go-linux-amd64 build pbf2json.go;
chmod +x pbf2json;
mv pbf2json build/pbf2json.linux-x64;

# echo "[compile] darwin i386";
# go-darwin-386 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.darwin-ia32;

echo "[compile] darwin x64";
go-darwin-amd64 build pbf2json.go;
chmod +x pbf2json;
mv pbf2json build/pbf2json.darwin-x64;

# echo "[compile] windows i386";
# go-windows-386 build pbf2json.go;
# chmod +x pbf2json.exe;
# mv pbf2json.exe build/pbf2json.win32-ia32;

echo "[compile] windows x64";
go-windows-amd64 build pbf2json.go;
chmod +x pbf2json.exe;
mv pbf2json.exe build/pbf2json.win32-x64;

# echo "[compile] freebsd arm";
# go-freebsd-arm build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-arm;

# echo "[compile] freebsd i386";
# go-freebsd-386 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-ia32;

# echo "[compile] freebsd x64";
# go-freebsd-amd64 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.freebsd-x64;

# echo "[compile] openbsd i386";
# go-openbsd-386 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.openbsd-ia32;

# echo "[compile] openbsd x64";
# go-openbsd-amd64 build pbf2json.go;
# chmod +x pbf2json;
# mv pbf2json build/pbf2json.openbsd-x64;