#!/bin/sh
set -e

if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' is niet geïnstalleerd. Installeer Go eerst voordat je dit script uitvoert."
    exit 1
fi

# 1. Download source and build for Linux
echo "Downloading and building somtoday-cli for Linux..."
git clone https://github.com/poorqualitycat/somtoday-cli.git /tmp/somtoday-cli-src
cd /tmp/somtoday-cli-src
go build -o somtoday-cli
mv somtoday-cli ../somtoday-cli
cd ..
rm -rf /tmp/somtoday-cli-src

# 2. Make it executable
chmod +x somtoday-cli

# 3. Move it to the system PATH (requires sudo)
echo "Installing to /usr/local/bin..."
sudo mv somtoday-cli /usr/local/bin/somtoday-cli

# 4. Create symlink for TUI
sudo ln -sf /usr/local/bin/somtoday-cli /usr/local/bin/somtoday-tui

echo "\033[0;32mSuccessfully installed!\033[0m"
echo "Probeer de CLI: somtoday-cli help"
echo "Of start de TUI: somtoday-tui"
