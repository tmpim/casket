#!/usr/bin/env bash
# Casket Automatic Install Script
# https://github.com/tmpim/casket

requires() {
    if ! [ -x "$(command -v $1)" ]; then
        echo "Error: $1 is not installed. It is required to run this script." >&2
        exit 1
    fi
}

# Prerequisite programs
requires curl
requires jq
requires mktemp
requires uname

uname_m=$(uname -m)
if [[ $uname_m == *64* ]]; then
	casket_arch="amd64"
elif [[ $uname_m == *86* ]]; then
    casket_arch="386"
fi

# Defaults for Linux and macOS
casket_ext=".tar.gz"
casket_bin="casket"
default_install_path="/usr/local/bin"
sudo_cmd="sudo"

uname_full=$(uname)
uname_upper=${uname_full^^}
if [[ $uname_upper == *LINUX* ]]; then
    casket_os="linux"
    requires tar
elif [[ $uname_upper == *DARWIN* ]]; then
	casket_os="darwin"
    requires tar
elif [[ $uname_upper == *WINDOWS* ]]; then
    casket_os="windows"
    casket_ext=".zip"
    casket_bin="casket.exe"
    sudo_cmd=""
    requires unzip
fi

if [ -z ${INSTALL_PATH+x} ]; then
    INSTALL_PATH=$default_install_path
    echo "INSTALL_PATH not specified, installing to $default_install_path"
fi

if [[ $casket_os == darwin && $casket_arch == 386 ]]; then
    echo "Error: Casket is not supported in 32-bit macOS."
    exit 1
fi

echo "Identified platform as ${casket_os}_${casket_arch}"
echo "Fetching latest available release..."

casket_tag=$(curl --silent "https://api.github.com/repos/tmpim/casket/releases/latest" | jq -r ".tag_name")

echo "Latest Casket release is ${casket_tag}"

casket_version=${casket_tag:1}
casket_file="casket_${casket_version}_${casket_os}_${casket_arch}${casket_ext}"
casket_url="https://github.com/tmpim/casket/releases/download/${casket_tag}/${casket_file}"

casket_tmp=$(mktemp -d)
casket_dl="${casket_tmp}/${casket_file}"

echo "Downloading Casket from $casket_url"
echo "Saving to $casket_dl"

curl -fsSL "$casket_url" -o "$casket_dl"

echo "Extracting..."
case "$casket_ext" in
    .zip)    unzip -o "$casket_dl" "$casket_bin" -d "$casket_tmp" ;;
    .tar.gz) tar -xzf "$casket_dl" -C "$casket_tmp" "$casket_bin" ;;
esac

casket_bin_dl="${casket_tmp}/${casket_bin}"
echo "Setting binary as executable..."
chmod +x "$casket_bin_dl"

casket_bin_install="$INSTALL_PATH/$casket_bin"
echo "Moving $casket_bin_dl to $casket_bin_install"
$sudo_cmd mv "$casket_bin_dl" "$casket_bin_install"

echo "Removing $casket_dl"
rm -- "$casket_dl"

if ! [ -x "$(command -v setcap)" ]; then
    echo "Setting bind capabilities on $casket_bin_install"
    $sudo_cmd setcap cap_net_bind_service=+ep "$casket_bin_install"
fi

$casket_bin_install -version
echo "Successfully installed Casket, welcome to the future!"