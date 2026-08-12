# 1. Zorg dat het script direct stopt bij een fout
$ErrorActionPreference = "Stop"

if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "Error: 'go' is niet geïnstalleerd. Installeer Go (https://go.dev/dl/) voordat je dit script uitvoert." -ForegroundColor Red
    exit 1
}

# 2. Definieer de installatiemap in Program Files
$InstallDir = "$env:ProgramFiles\somtoday-cli"
$ExePath = "$InstallDir\somtoday-cli.exe"
$TuiPath = "$InstallDir\somtoday-tui.exe"

# 3. Maak de map aan als deze nog niet bestaat
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

# 4. Download de broncode en bouw voor Windows
Write-Host "Downloading and building somtoday-cli for Windows..." -ForegroundColor Cyan
$TempDir = Join-Path $env:TEMP "somtoday-cli-src"
if (Test-Path $TempDir) {
    Remove-Item -Recurse -Force $TempDir
}
git clone https://github.com/poorqualitycat/somtoday-cli.git $TempDir
Push-Location $TempDir
go build -o $ExePath
Pop-Location
Remove-Item -Recurse -Force $TempDir

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
