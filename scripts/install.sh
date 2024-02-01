#!/usr/bin/env sh
set -u

if [[ ${OS:-} = Windows_NT ]]; then
    echo 'error: If you are a Windows user, please install in WSL (Windows Subsystem for Linux).'
    exit 1
fi

BOLD="$(tput bold 2>/dev/null || printf '')"
GREY="$(tput setaf 0 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
NO_COLOR="$(tput sgr0 2>/dev/null || printf '')"

base_path=$(pwd)
version=v2.0.0
platform="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
case "${arch}" in
  x86_64) arch="amd64" ;;
  aarch64) arch="arm64" ;;
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

download_release() {
  out=$1
  file=$2

  url="https://github.com/squarecloudofc/cli/releases/download/$version/${file}"
  
  if has curl; then
    curl -sL -o "${out}" "${url}"
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

execute() {
  tmpdir=$(mktemp -d)

  filename="squarecloud_${platform}_${arch}.tar.gz"

  info "Installing Square Cloud CLI, please wait..."
  download_release "${tmpdir}/${filename}" $filename

  info "Unpacking ${filename}"
  untar "${tmpdir}/${filename}" "${tmpdir}/bin"

  sudo cp -f "${tmpdir}/bin/squarecloud" "/usr/local/bin"
  completed "Successfuly installed Square Cloud CLI!"

  rm -rf "${tmpdir}"
}

execute
