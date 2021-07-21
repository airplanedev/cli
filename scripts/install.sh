#!/bin/sh

# Inspired by:
# - https://github.com/denoland/deno_install/blob/master/install.sh
#   Copyright 2019 the Deno authors. All rights reserved. MIT license.
# - https://github.com/superfly/flyctl/blob/master/installers/install.sh

set -e

case $(uname -sm) in
    "Darwin x86_64") target="darwin_x86_64" ;;
    "Darwin arm64") target="darwin_arm64" ;;
    *) target="linux_x86_64" ;;
esac

if [ $# -eq 0 ]; then
    download_uri="https://github.com/airplanedev/cli/releases/latest/download/airplane_${target}.tar.gz"
else
    download_uri="https://github.com/airplanedev/cli/releases/download/${1}/airplane_${target}.tar.gz"
fi

airplane_install="${AIRPLANE_INSTALL:-$HOME/.airplane}"
bin_dir="$airplane_install/bin"
exe="$bin_dir/airplane"

if [ ! -d "$bin_dir" ]; then
    mkdir -p "$bin_dir"
fi

curl --fail --location --progress-bar --output "$exe.tar.gz" "$download_uri"
cd "$bin_dir"
tar xzf "$exe.tar.gz"
chmod +x "$exe"
rm "$exe.tar.gz"

echo "The Airplane CLI was installed successfully to $exe"
if command -v airplane >/dev/null; then
    echo "Run 'airplane --help' to get started."
else
    case $SHELL in
    /bin/zsh) shell_profile=".zshrc" ;;
    *) shell_profile=".bash_profile" ;;
    esac
    echo "✋ Manually add the following to your \$HOME/$shell_profile (or similar):"
    echo "    export AIRPLANE_INSTALL=\"$airplane_install\""
    echo "    export PATH=\"\$AIRPLANE_INSTALL/bin:\$PATH\""
    echo ""
    echo "Then, run 'airplane --help' to get started."
fi
