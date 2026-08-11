#!/bin/sh
set -e

# 1. Bepaal de architectuur
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    BINARY_URL="https://github.com/poorqualitycat/somtoday-cli/releases/latest/download/somtoday-darwin-arm64"
else
    BINARY_URL="https://github.com/poorqualitycat/somtoday-cli/releases/latest/download/somtoday-darwin-amd64"
fi

# 2. Download de juiste macOS binary naar een tijdelijke map
echo "Downloading somtoday-cli for macOS ($ARCH)..."
curl -sSfL "$BINARY_URL" -o /tmp/somtoday-cli || {
    echo "Release niet gevonden. Lokaal bouwen..."
    git clone https://github.com/poorqualitycat/somtoday-cli.git /tmp/somtoday-cli-src
    cd /tmp/somtoday-cli-src
    go build -o /tmp/somtoday-cli
    cd ..
    rm -rf /tmp/somtoday-cli-src
}

# 3. Maak het bestand uitvoerbaar
chmod +x /tmp/somtoday-cli

# 4. Zorg dat de doelmap bestaat
mkdir -p /usr/local/bin

# 5. Verplaats de binary naar de systeem-PATH (vraagt eenmalig om je Mac-wachtwoord)
echo "Installing to /usr/local/bin..."
sudo mv /tmp/somtoday-cli /usr/local/bin/somtoday-cli

# 6. Maak de TUI symlink aan
sudo ln -sf /usr/local/bin/somtoday-cli /usr/local/bin/somtoday-tui

# 7. macOS Gatekeeper fix
sudo xattr -r -d com.apple.quarantine /usr/local/bin/somtoday-cli 2>/dev/null || true

echo "\033[0;32mSuccessfully installed!\033[0m"
echo "Probeer de CLI: somtoday-cli help"
echo "Of start de TUI: somtoday-tui"
