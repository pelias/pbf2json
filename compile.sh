#!/bin/bash

# Cross-compiles the Go binaries shipped in the npm package.
#
# Called with no arguments it builds every supported target, which is what the
# 'prepack' lifecycle script does when the package is published. Pass 'native'
# (or explicit target names) to build a subset, eg. './compile.sh native'.
#
# note: to cross-compile on your own machine you may need to follow the guide at
# http://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5

set -o pipefail

# 'go' is not always on the PATH in CI, fall back to the runner tool cache
if ! command -v go &> /dev/null; then
  for candidate in "${RUNNER_TOOL_CACHE}"/go/*/x64/bin /usr/local/go/bin; do
    if [[ -x "${candidate}/go" ]]; then
      export PATH="${candidate}:${PATH}"
      break
    fi
  done
fi

if ! command -v go &> /dev/null; then
  echo "the go compiler is required to build pbf2json" >&2
  exit 1
fi

# target name -> GOOS GOARCH 'expected file(1) output' use_upx
declare -A TARGETS=(
  ['linux-x64']="linux amd64 'ELF 64-bit LSB.*x86-64' yes"
  ['linux-arm64']="linux arm64 'ELF 64-bit LSB.*aarch64' yes"
  ['darwin-x64']="darwin amd64 'Mach-O 64-bit.*x86_64' no"
  ['darwin-arm64']="darwin arm64 'Mach-O 64-bit.*arm64' no"
  # UPX is disabled on darwin due to https://github.com/upx/upx/issues/187
  ['win32-x64']="windows amd64 'PE32\+ executable.*x86-64' yes"
)

# the target matching the machine we're running on
function native_target() {
  local goos goarch
  goos=$(go env GOOS)
  goarch=$(go env GOARCH)
  [[ "${goos}" == 'windows' ]] && goos='win32'
  [[ "${goarch}" == 'amd64' ]] && goarch='x64'
  echo "${goos}-${goarch}"
}

function build() {
  local name="$1"
  if [[ -z "${TARGETS[$name]}" ]]; then
    echo "unknown compile target: ${name}" >&2
    exit 1
  fi

  local goos goarch expected upx
  eval "set -- ${TARGETS[$name]}"
  goos="$1"; goarch="$2"; expected="$3"; upx="$4"

  local out="build/pbf2json.${name}"

  echo "[compile] ${name}";
  env GOOS="${goos}" GOARCH="${goarch}" go build -ldflags="-s -w" \
    -gcflags=-trimpath="${GOPATH}" -asmflags=-trimpath="${GOPATH}" -o "${out}"
  if [[ $? != 0 ]]; then
    echo "failed to compile ${name}" >&2
    exit 1
  fi
  chmod +x "${out}"

  # confirm the binary is actually for the architecture we asked for
  local actual
  actual=$(file -b "${out}")
  if ! grep -qE "${expected}" <<< "${actual}"; then
    echo "invalid file architecture: ${out}" >&2
    echo "expected: ${expected}" >&2
    echo "actual: ${actual}" >&2
    exit 1
  fi

  # if the 'UPX' binary packer is available, use it https://upx.github.io/
  if [[ "${upx}" == 'yes' ]] && command -v upx &> /dev/null; then
    upx "${out}"
  fi
}

targets=("$@")
if [[ ${#targets[@]} == 0 ]]; then
  targets=("${!TARGETS[@]}")
  rm -rf build # a full build starts from scratch, partial builds do not
elif [[ "${targets[0]}" == 'native' ]]; then
  targets=("$(native_target)")
fi

mkdir -p build
for target in "${targets[@]}"; do
  build "${target}"
done
