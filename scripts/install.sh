#!/usr/bin/env sh
set -u

if [ "${OS:-}" = "Windows_NT" ]; then
    echo 'error: If you are a Windows user, please install in WSL (Windows Subsystem for Linux).'
    exit 1
fi

BOLD="$(tput bold 2>/dev/null || printf '')"
GREY="$(tput setaf 0 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
NO_COLOR="$(tput sgr0 2>/dev/null || printf '')"

REPO='squarecloudofc/cli'

PLATFORM="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | tr '[:upper:]' '[:lower:]')"
case "${ARCH}" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv7*) ARCH="armv7" ;;
  i386 | i686) ARCH="386" ;;
esac

has() {
  command -v "$1" 1>/dev/null 2>&1
}

info() {
  printf '%s\n' "${BOLD}${GREY}>${NO_COLOR} $*"
}

warn() {
  printf '%s\n' "${YELLOW}! $*${NO_COLOR}"
}

error() {
  printf '%s\n' "${RED}x $*${NO_COLOR}" >&2
}

success() {
  printf '%s\n' "${GREEN}$@ ${NO_COLOR}"
}

completed() {
  printf '%s\n' "${GREEN}✓${NO_COLOR} $*"
}

get_latest_tag(){
  url="https://api.github.com/repos/${REPO}/releases/latest"

  if has curl; then
    curl -sL $url | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
  elif has wget; then
    wget -q -O- $url | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
  else
    error "Unable to find a latest release in github"
    return 1
  fi
}

VERSION=$(get_latest_tag)
if [ -z "${VERSION}" ]; then
  error "Unable to find the latest release (GitHub API unreachable or rate limited)"
  exit 1
fi

download_release() {
  out=$1
  file=$2

  url="https://github.com/${REPO}/releases/download/$VERSION/${file}"

  if has curl; then
    curl -fsSL -o "${out}" "${url}"
  elif has wget; then
    wget -q -O "${out}" "${url}"
  else
    error "Unable to find a HTTP download program"
    return 1
  fi
}

untar() {
  file=$1
  outdir=$2

  if [ ! -d $outdir ]; then
    mkdir -p $outdir
  fi

  case "${file}" in
    *.tar.gz | *.tgz) tar --no-same-owner -xzf "${file}" -C "${outdir}" ;;
    *.tar) tar --no-same-owner -xf "${file}" -C "${outdir}" ;;
    *.zip) unzip "${file}" -d "${outdir}" ;;
    *)
      error "untar unknown archive format for ${file}"
      return 1
      ;;
  esac
}

verify_checksum() {
  # Not "file"/"out": sh has no locals and download_release sets those.
  archive=$1
  name=$2

  download_release "${tmpdir}/checksums.txt" checksums.txt || return 1
  expected=$(awk -v f="${name}" '$2 == f { print $1 }' "${tmpdir}/checksums.txt")

  if has sha256sum; then
    actual=$(sha256sum "${archive}" | awk '{ print $1 }')
  elif has shasum; then
    actual=$(shasum -a 256 "${archive}" | awk '{ print $1 }')
  else
    error "sha256sum or shasum is required to verify the download"
    return 1
  fi

  [ -n "${expected}" ] && [ "${actual}" = "${expected}" ]
}

fail() {
  error "$*"
  rm -rf "${tmpdir}"
  exit 1
}

execute() {
  tmpdir=$(mktemp -d) || exit 1

  filename="squarecloud_${PLATFORM}_${ARCH}.tar.gz"

  info "Installing Square Cloud CLI, please wait..."
  download_release "${tmpdir}/${filename}" "${filename}" || fail "Unable to download ${filename} (unsupported platform or architecture?)"
  verify_checksum "${tmpdir}/${filename}" "${filename}" || fail "Checksum verification failed for ${filename}"

  info "Unpacking ${filename}"
  untar "${tmpdir}/${filename}" "${tmpdir}/bin" || fail "Unable to unpack ${filename}"

  sudo=""
  if [ "$(id -u)" != "0" ]; then
    sudo="sudo"
  fi
  ${sudo} cp -f "${tmpdir}/bin/squarecloud" "/usr/local/bin" || fail "Unable to copy squarecloud to /usr/local/bin"
  completed "Successfully installed Square Cloud CLI $VERSION!"

  rm -rf "${tmpdir}"
}

execute
