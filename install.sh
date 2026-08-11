#!/bin/sh
set -e

# 1. Download your pre-compiled Linux binary
echo "Downloading somtoday-cli for Linux..."
# Voor nu compileren we hem lokaal ter demonstratie als we git gebruiken, anders halen we een dummy release op:
curl -sL "https://github.com/poorqualitycat/somtoday-cli/releases/latest/download/somtoday-linux-amd64" -o somtoday-cli || {
    echo "Release niet gevonden. Gebruik fallback (lokaal bouwen)..."
    git clone https://github.com/poorqualitycat/somtoday-cli.git /tmp/somtoday-cli-src
    cd /tmp/somtoday-cli-src
    go build -o somtoday-cli
    mv somtoday-cli ../somtoday-cli
    cd ..
    rm -rf /tmp/somtoday-cli-src
}

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
