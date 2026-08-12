#!/bin/sh
set -e

if ! command -v go >/dev/null 2>&1; then
    echo "Error: 'go' is niet geïnstalleerd. Installeer Go (bijv. via 'brew install go') voordat je dit script uitvoert."
    exit 1
fi

# 1. Download source and build for macOS
echo "Downloading and building somtoday-cli for macOS..."
git clone https://github.com/poorqualitycat/somtoday-cli.git /tmp/somtoday-cli-src
cd /tmp/somtoday-cli-src
go build -o /tmp/somtoday-cli
cd ..
rm -rf /tmp/somtoday-cli-src

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
