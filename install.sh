#!/bin/sh

set -eu

repository="Hayao0819/nth"

fail() {
    printf 'nth: %s\n' "$1" >&2
    exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"

case "$(uname -s)" in
    Linux)
        os="linux"
        ;;
    Darwin)
        os="darwin"
        ;;
    *)
        fail "unsupported operating system: $(uname -s)"
        ;;
esac

case "$(uname -m)" in
    x86_64 | amd64)
        arch="amd64"
        ;;
    arm64 | aarch64)
        arch="arm64"
        ;;
    *)
        fail "unsupported architecture: $(uname -m)"
        ;;
esac

latest_url=$(curl -fsSL --proto '=https' --tlsv1.2 -o /dev/null \
    -w '%{url_effective}' "https://github.com/$repository/releases/latest")
latest_url=${latest_url%/}
tag=${latest_url##*/}

case "$tag" in
    v[0-9]* | [0-9]*)
        ;;
    *)
        fail "could not determine the latest release tag"
        ;;
esac

version=${tag#v}
asset="nth_${version}_${os}_${arch}"
release_url="https://github.com/$repository/releases/download/$tag"

if [ -n "${NTH_INSTALL_DIR:-}" ]; then
    install_dir=$NTH_INSTALL_DIR
else
    [ -n "${HOME:-}" ] || fail "HOME is not set"
    install_dir="$HOME/.local/bin"
fi

mkdir -p "$install_dir"
temporary_binary=""
temporary_checksums=""

cleanup() {
    if [ -n "${temporary_binary:-}" ]; then
        rm -f "$temporary_binary"
    fi
    if [ -n "${temporary_checksums:-}" ]; then
        rm -f "$temporary_checksums"
    fi
}
trap cleanup 0
trap 'exit 1' HUP INT TERM

temporary_binary=$(mktemp "$install_dir/.nth.XXXXXX")
temporary_checksums=$(mktemp "$install_dir/.nth-checksums.XXXXXX")

curl -fsSL --retry 3 --proto '=https' --tlsv1.2 \
    -o "$temporary_binary" "$release_url/$asset"
curl -fsSL --retry 3 --proto '=https' --tlsv1.2 \
    -o "$temporary_checksums" "$release_url/checksums.txt"

expected_checksum=$(awk -v asset="$asset" '$2 == asset { print $1; exit }' "$temporary_checksums")
[ -n "$expected_checksum" ] || fail "checksum for $asset was not found"

if command -v sha256sum >/dev/null 2>&1; then
    actual_checksum=$(sha256sum "$temporary_binary")
elif command -v shasum >/dev/null 2>&1; then
    actual_checksum=$(shasum -a 256 "$temporary_binary")
else
    fail "sha256sum or shasum is required"
fi
actual_checksum=${actual_checksum%% *}

[ "$actual_checksum" = "$expected_checksum" ] || fail "checksum verification failed"

chmod 0755 "$temporary_binary"
target="$install_dir/nth"
mv -f "$temporary_binary" "$target"
temporary_binary=""

printf 'Installed nth %s to %s\n' "$tag" "$target"
case ":${PATH:-}:" in
    *":$install_dir:"*)
        ;;
    *)
        printf 'Add %s to PATH to run nth.\n' "$install_dir" >&2
        ;;
esac
