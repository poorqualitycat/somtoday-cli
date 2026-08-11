# 1. Zorg dat het script direct stopt bij een fout
$ErrorActionPreference = "Stop"

# 2. Definieer de installatiemap in Program Files
$InstallDir = "$env:ProgramFiles\somtoday-cli"
$ExePath = "$InstallDir\somtoday-cli.exe"
$TuiPath = "$InstallDir\somtoday-tui.exe"

# 3. Maak de map aan als deze nog niet bestaat
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

# 4. Download de gecompileerde Windows binary van jouw release
Write-Host "Downloading somtoday-cli for Windows..." -ForegroundColor Cyan
$DownloadUrl = "https://github.com/poorqualitycat/somtoday-cli/releases/latest/download/somtoday-windows-amd64.exe"
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ExePath
} catch {
    Write-Host "Release niet gevonden. (Je zult lokaal moeten bouwen via 'go build'.)" -ForegroundColor Red
    exit 1
}

# Maak een kopie voor het TUI commando (zodat somtoday-tui hetzelfde bestand start)
Copy-Item -Path $ExePath -Destination $TuiPath -Force

# 5. Voeg de map toe aan de Systeem PATH zodat het commando overal werkt
Write-Host "Adding to System PATH..." -ForegroundColor Cyan
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($CurrentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", $CurrentPath + ";$InstallDir", "Machine")
    # Vernieuw de PATH direct in de huidige terminal-sessie
    $env:Path += ";$InstallDir"
}

Write-Host "Successfully installed!" -ForegroundColor Green
Write-Host "Restart your terminal and try running:" -ForegroundColor Yellow
Write-Host "  somtoday-cli help" -ForegroundColor Yellow
Write-Host "  somtoday-tui" -ForegroundColor Yellow
