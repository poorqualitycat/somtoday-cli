# Somtoday CLI & TUI (Go)

Een moderne, razendsnelle Command Line Interface (CLI) en Terminal User Interface (TUI) voor **Somtoday**, geschreven in Go.

![Somtoday](https://img.shields.io/badge/Somtoday-CLI%20%26%20TUI-E02475?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.20+-00ADD8?style=for-the-badge&logo=go)

---

## 🚀 Installatie

Installeer `somtoday-cli` en `somtoday-tui` direct via de terminal:

### 🌐 1. Automatische Installers (Aanbevolen)

#### Linux
```bash
curl -sSfL https://raw.githubusercontent.com/poorqualitycat/somtoday-cli/refs/heads/main/install.sh | sh
```

#### macOS
```bash
curl -sSfL https://raw.githubusercontent.com/poorqualitycat/somtoday-cli/refs/heads/main/install-mac.sh | sh
```

#### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/poorqualitycat/somtoday-cli/refs/heads/main/install.ps1 | iex
```

---

### 🛠️ 2. Handmatig bouwen (Vanaf broncode)

Als je Go geïnstalleerd hebt op je systeem:

```bash
# Repository kloonen
git clone https://github.com/poorqualitycat/somtoday-cli.git
cd somtoday-cli

# Binary compileren
go build -o somtoday-cli

# (Optioneel) Naar je PATH verplaatsen en TUI koppelen op Linux/Mac
sudo mv somtoday-cli /usr/local/bin/somtoday-cli
sudo ln -sf /usr/local/bin/somtoday-cli /usr/local/bin/somtoday-tui
```

---

## 💻 Gebruik

Het programma heeft twee standen: **TUI** (interactieve visuele modus) en **CLI** (snelle commando's in je terminal).

### 🎨 1. Terminal User Interface (TUI)
Start de interactieve GUI in Somtoday merkkleuren:
```bash
somtoday-tui
```

### ⚡ 2. Command Line Interface (CLI)
Voer snelle commando's uit zonder de TUI te openen:

#### Inloggen
```bash
# Inloggen via SSO (geeft een eenmalige browser-link, daarna blijf je altijd ingelogd)
somtoday-cli login --school "Mijn School" --sso

# Of inloggen met gebruikersnaam en wachtwoord
somtoday-cli login --school "Mijn School" --username "gebruiker" --password "wachtwoord"
```

#### Cijfers
```bash
# Alle cijfers bekijken
somtoday-cli cijfers

# Filteren op vak en sortering
somtoday-cli cijfers --vak "wiskunde" --sort "hoog/laag"
somtoday-cli cijfers --sort "laag/hoog"
```

#### Planning / Rooster
```bash
# Planning van vandaag
somtoday-cli planning

# Specifieke opties
somtoday-cli planning --huidige
somtoday-cli planning --eerstvolgende
somtoday-cli planning --tijd "09:00"
somtoday-cli planning --dag "maandag" --week "36" --jaar "2026"
```

#### Huiswerk
```bash
# Huiswerk bekijken
somtoday-cli huiswerk

# Filteren op voltooid / onvoltooid
somtoday-cli huiswerk --onvoltooid
somtoday-cli huiswerk --voltooid
somtoday-cli huiswerk --beschrijving "Engels"
```

#### Overige Commando's
```bash
somtoday-cli afwezigheid   # Toont alle afwezigheden van het schooljaar
somtoday-cli school-info  # Toont school- en accountinformatie
somtoday-cli berichten    # Toont recente berichten
somtoday-cli logout       # Wist de opgeslagen configuratie
somtoday-cli help         # Toont het overzicht van alle commando's
```

---

## 🔒 Beveiliging & Opslag

De applicatie maakt gebruik van een automatische `config.json` opslag in OS-specifieke mappen:
- **Linux:** `~/.config/somtoday-cli/config.json`
- **Windows:** `%APPDATA%\somtoday-cli\config.json`
- **macOS:** `~/Library/Preferences/somtoday-cli/config.json`

Zodra je via SSO of Wachtwoord inlogt, wordt er een **OAuth Refresh Token** opgeslagen. Hierdoor blijf je **altijd ingelogd**, zelfs als de CLI opnieuw opstart of je pc herstart wordt!

---

## 📖 SOMtoday REST API Documentatie

Hieronder vind je de originele documentatie van de Somtoday REST API.

## Table of contents

- [Authentication / authorization](Authentication.md)
- [Homework](Homework.md)
- [Fetching information](#fetching-information)
  - [Current student(s): `GET /rest/v1/leerlingen`](#current-students-get-restv1leerlingen)
  - [Grades: `GET /rest/v1/resultaten/huidigVoorLeerling/[id]`](#grades-get-restv1resultatenhuidigvoorleerlingid)
  - [Schedule: `GET /rest/v1/afspraken`](#schedule-get-restv1afspraken)
  - [Absence Reports: `GET /rest/v1/absentiemeldingen`](#absence-reports-get-restv1absentiemeldingen)
