package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Kleuren (SomToday branding) ──────────────────────────────────────────────

const (
	kleurPrimair    = "#66B4FF" // Highlighted text / Active
	kleurSecundair  = "#0084FF" // Somtoday Electric Blue
	kleurBasis      = "#121214" // Main console background
	kleurFiller     = "#252529" // Filler color
	kleurTekst      = "#F5F5F7" // Text color
	
	kleurZeerVoldoende = "#34C759" // Zeer voldoende (7.5 - 10)
	kleurVoldoende     = "#F5F5F7" // Voldoende (5.5 - 7.4)
	kleurOnvoldoende   = "#FF3B30" // Onvoldoende (1.0 - 5.4)

	kleurGroen      = "#34C759" // Afgevinkt
	kleurRood       = "#FF3B30" // Uitval
	kleurOranje     = "#F97316" // Te laat
	kleurWit        = "#FFFFFF"
	kleurLichtGrijs = "#F4F5F8" 
	kleurGrijs      = "#94A3B8"
	kleurZwart      = "#1E293B" 
	kleurSubtiel    = "#64748B"
)

// ─── Views ────────────────────────────────────────────────────────────────────

type schermType int

const (
	schermLaden schermType = iota
	schermSchoolZoeken
	schermInlogKeuze
	schermBrowserLogin
	schermInloggen
	schermHoofdmenu
	schermRooster
	schermCijfers
	schermAbsenties
	schermHuiswerk
	schermProfiel
	schermFout
)

// ─── Model ────────────────────────────────────────────────────────────────────

type model struct {
	sessie      *Sessie
	scherm      schermType
	breedte     int
	hoogte      int
	fout        error
	statusTekst string

	// School zoeken
	scholen         []School
	schoolFilter    string
	schoolSelectie  int
	scholenGeladen  bool

	// Inloggen
	inlogKeuze      int       // 0=Wachtwoord, 1=Browser
	inlogVelden     [3]string // 0=school(readonly), 1=gebruikersnaam, 2=wachtwoord
	inlogFocus      int       // welk veld actief is (1 of 2)
	inlogBezig      bool
	gekozenSchool   *School

	// Browser Login
	browserURL      string
	browserVerifier string
	browserInput    string

	// Data
	leerling   *Leerling
	cijfers    []Cijfer
	rooster    []Afspraak
	absenties  []Absentie
	huiswerk   []Huiswerk

	// Menu
	menuSelectie int
	laden        bool

	// Scrolling
	scrollOffset int

	uitloggen bool
}

// ─── Berichten ────────────────────────────────────────────────────────────────

type scholenMsg struct {
	scholen []School
	err     error
}

type inlogMsg struct {
	err error
}
type ipcMsg string

type leerlingMsg struct {
	leerling *Leerling
	err      error
}

type cijfersMsg struct {
	cijfers []Cijfer
	err     error
}

type roosterMsg struct {
	rooster []Afspraak
	err     error
}

type absentiesMsg struct {
	absenties []Absentie
	err       error
}

type huiswerkMsg struct {
	huiswerk []Huiswerk
	err      error
}

type tokenVernieuwdMsg struct {
	err error
}

// ─── Styles ───────────────────────────────────────────────────────────────────

var (
	stijlHeader = lipgloss.NewStyle().
		Background(lipgloss.Color(kleurBasis)).
		Foreground(lipgloss.Color(kleurWit)).
		Bold(true).
		Padding(0, 2)

	stijlHeaderInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurSubtiel)).
		PaddingLeft(1)

	stijlKader = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(kleurBasis)).
		Padding(1, 2)

	stijlKaderActief = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(kleurPrimair)).
		Padding(1, 2)

	stijlMenu = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurTekst))

	stijlMenuActief = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurPrimair)).
		Bold(true)

	stijlLabel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurPrimair)).
		Bold(true)

	stijlWaarde = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurTekst))

	stijlSubtiel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurSubtiel))

	stijlFout = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurRood)).
		Bold(true)

	stijlSucces = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurGroen)).
		Bold(true)

	stijlCijferZeerVoldoende = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurZeerVoldoende)).
		Bold(true)
		
	stijlCijferVoldoende = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurVoldoende)).
		Bold(true)

	stijlCijferOnvoldoende = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurOnvoldoende)).
		Bold(true)

	stijlOnvoldoende = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurOnvoldoende)).
		Bold(true)

	stijlVoet = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurSubtiel)).
		MarginTop(1)

	stijlTitel = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurPrimair)).
		Bold(true).
		MarginBottom(1)

	stijlInputActief = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(kleurPrimair)).
		Padding(0, 1).
		MarginBottom(1)

	stijlInputInactief = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(kleurGrijs)).
		Padding(0, 1).
		MarginBottom(1)

	stijlZoekbalk = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(kleurPrimair)).
		Padding(0, 1).
		MarginBottom(1)

	stijlSchoolItem = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurTekst)).
		PaddingLeft(2)

	stijlSchoolItemActief = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurPrimair)).
		Bold(true).
		PaddingLeft(1)

	stijlTabelHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurWit)).
		Background(lipgloss.Color(kleurBasis)).
		Bold(true).
		Padding(0, 1)

	stijlRij = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurTekst)).
		Padding(0, 1)

	stijlUitval = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurRood)).
		Bold(true)

	stijlHuidig = lipgloss.NewStyle().
		Foreground(lipgloss.Color(kleurOranje)).
		Bold(true)
)

// ─── Init ─────────────────────────────────────────────────────────────────────

func startTUI() model {
	p := tea.NewProgram(beginModel(), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Fout bij starten: %v\n", err)
		os.Exit(1)
	}
	return m.(model)
}

func beginModel() model {
	sessie := nieuweSessie()
	m := model{
		sessie:     sessie,
		scherm:     schermLaden,
		laden:      true,
		inlogFocus: 1,
	}
	return m
}

func (m model) Init() tea.Cmd {
	sessie := m.sessie
	if sessie.isIngelogd() {
		// Probeer token te vernieuwen
		return func() tea.Msg {
			err := sessie.tokenVernieuwen()
			return tokenVernieuwdMsg{err: err}
		}
	}
	// Begin met een lege lijst scholen (zoeken op aanvraag)
	return func() tea.Msg {
		return scholenMsg{scholen: []School{}, err: nil}
	}
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.breedte = msg.Width
		m.hoogte = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.toetsAfhandelen(msg)

	case scholenMsg:
		return m.scholenOntvangen(msg)

	case inlogMsg:
		return m.inlogResultaat(msg)

	case ipcMsg:
		m.browserInput = string(msg)
		m.inlogBezig = true
		schoolUUID := m.gekozenSchool.UUID
		verifier := m.browserVerifier
		input := m.browserInput
		sessie := m.sessie
		return m, func() tea.Msg {
			err := sessie.voltooiBrowserLogin(schoolUUID, verifier, input)
			return inlogMsg{err: err}
		}

	case tokenVernieuwdMsg:
		return m.tokenVernieuwdAfhandelen(msg)

	case leerlingMsg:
		m.laden = false
		if msg.err != nil {
			m.fout = msg.err
			m.scherm = schermFout
			return m, nil
		}
		m.leerling = msg.leerling
		m.scherm = schermHoofdmenu
		return m, nil

	case cijfersMsg:
		m.laden = false
		if msg.err != nil {
			m.fout = msg.err
			m.scherm = schermFout
			return m, nil
		}
		m.cijfers = msg.cijfers
		m.scherm = schermCijfers
		m.scrollOffset = 0
		return m, nil

	case roosterMsg:
		m.laden = false
		if msg.err != nil {
			m.fout = msg.err
			m.scherm = schermFout
			return m, nil
		}
		m.rooster = msg.rooster
		m.scherm = schermRooster
		m.scrollOffset = 0
		return m, nil

	case absentiesMsg:
		m.laden = false
		if msg.err != nil {
			m.fout = msg.err
			m.scherm = schermFout
			return m, nil
		}
		m.absenties = msg.absenties
		m.scherm = schermAbsenties
		m.scrollOffset = 0
		return m, nil

	case huiswerkMsg:
		m.laden = false
		if msg.err != nil {
			m.fout = msg.err
			m.scherm = schermFout
			return m, nil
		}
		m.huiswerk = msg.huiswerk
		m.scherm = schermHuiswerk
		m.scrollOffset = 0
		return m, nil
	}

	return m, nil
}

func (m model) tokenVernieuwdAfhandelen(msg tokenVernieuwdMsg) (model, tea.Cmd) {
	if msg.err != nil {
		// Token vernieuwen mislukt → opnieuw inloggen
		m.sessie.AccessToken = ""
		m.sessie.RefreshToken = ""
		m.laden = false
		m.scherm = schermSchoolZoeken
		m.scholen = []School{}
		m.schoolFilter = ""
		return m, nil
	}
	// Token vernieuwd → leerling ophalen
	m.laden = true
	m.statusTekst = "Gegevens ophalen..."
	return m, func() tea.Msg {
		leerling, err := m.sessie.leerlingOphalen()
		return leerlingMsg{leerling: leerling, err: err}
	}
}

func (m model) scholenOntvangen(msg scholenMsg) (model, tea.Cmd) {
	m.laden = false
	if msg.err != nil {
		m.fout = msg.err
		m.scherm = schermFout
		return m, nil
	}
	m.scholen = msg.scholen
	m.scholenGeladen = true
	m.scherm = schermSchoolZoeken
	m.schoolSelectie = 0
	return m, nil
}

func (m model) inlogResultaat(msg inlogMsg) (model, tea.Cmd) {
	m.inlogBezig = false
	if msg.err != nil {
		m.fout = msg.err
		m.scherm = schermFout
		return m, nil
	}
	// Ingelogd! Leerling ophalen
	if m.gekozenSchool != nil {
		m.sessie.SchoolNaam = m.gekozenSchool.Naam
		configSet(keySchoolNaam, m.sessie.SchoolNaam)
	}
	m.laden = true
	m.statusTekst = "Gegevens ophalen..."
	m.scherm = schermLaden
	return m, func() tea.Msg {
		leerling, err := m.sessie.leerlingOphalen()
		return leerlingMsg{leerling: leerling, err: err}
	}
}

// ─── Toetsen ──────────────────────────────────────────────────────────────────

func (m model) toetsAfhandelen(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()

	// Globaal
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	}

	switch m.scherm {
	case schermSchoolZoeken:
		return m.toetsSchoolZoeken(msg)
	case schermInlogKeuze:
		return m.toetsInlogKeuze(msg)
	case schermBrowserLogin:
		return m.toetsBrowserLogin(msg)
	case schermInloggen:
		return m.toetsInloggen(msg)
	case schermHoofdmenu:
		return m.toetsMenu(msg)
	case schermCijfers, schermRooster, schermAbsenties, schermHuiswerk, schermProfiel:
		return m.toetsDataView(msg)
	case schermFout:
		switch key {
		case "esc", "enter":
			if m.sessie.isIngelogd() {
				m.scherm = schermHoofdmenu
				m.fout = nil
			} else {
				m.fout = nil
				if m.scholenGeladen {
					m.scherm = schermSchoolZoeken
				} else {
					m.scherm = schermSchoolZoeken
					m.laden = false
				}
			}
		case "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) toetsSchoolZoeken(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "q":
		return m, tea.Quit
	case "up", "ctrl+k":
		if m.schoolSelectie > 0 {
			m.schoolSelectie--
		}
	case "down", "ctrl+j":
		gefilterd := m.gefilterdeScholen()
		if m.schoolSelectie < len(gefilterd)-1 {
			m.schoolSelectie++
		}
	case "enter":
		gefilterd := m.gefilterdeScholen()
		if len(gefilterd) == 0 {
			if len(m.schoolFilter) >= 3 {
				m.statusTekst = "Zoeken..."
				m.scherm = schermLaden
				m.laden = true
				term := m.schoolFilter
				return m, func() tea.Msg {
					s, err := searchOrganisaties(term)
					return scholenMsg{scholen: s, err: err}
				}
			}
		} else if m.schoolSelectie < len(gefilterd) {
			school := gefilterd[m.schoolSelectie]
			m.gekozenSchool = &school
			m.inlogKeuze = 0
			m.scherm = schermInlogKeuze
		}
	case "backspace":
		if len(m.schoolFilter) > 0 {
			m.schoolFilter = m.schoolFilter[:len(m.schoolFilter)-1]
			m.schoolSelectie = 0
			m.scholen = []School{}
		}
	default:
		if len(key) == 1 {
			r := rune(key[0])
			if unicode.IsPrint(r) {
				m.schoolFilter += string(r)
				m.schoolSelectie = 0
				m.scholen = []School{}
			}
		} else if key == "space" {
			m.schoolFilter += " "
			m.schoolSelectie = 0
			m.scholen = []School{}
		}
	}
	return m, nil
}

func (m model) toetsInlogKeuze(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.scherm = schermSchoolZoeken
	case "up", "k":
		if m.inlogKeuze > 0 {
			m.inlogKeuze--
		}
	case "down", "j":
		if m.inlogKeuze < 1 {
			m.inlogKeuze++
		}
	case "enter":
		if m.inlogKeuze == 0 {
			m.inlogVelden[0] = m.gekozenSchool.Naam
			savedUser, _ := configGet("username")
			savedPass, _ := configGet("password")
			m.inlogVelden[1] = savedUser
			m.inlogVelden[2] = savedPass
			
			if savedUser != "" && savedPass != "" {
				// Automatisch inloggen starten als we gegevens hebben? Nee, laat de gebruiker bevestigen.
				m.inlogFocus = 2 // Focus op wachtwoord of gewoon op login knop? Focus op laatste veld
			} else {
				m.inlogFocus = 1 // Focus op gebruikersnaam
			}
			m.scherm = schermInloggen
		} else {
			authURL, verifier, err := startBrowserLogin(m.gekozenSchool.UUID)
			if err != nil {
				m.fout = err
				m.scherm = schermFout
			} else {
				m.browserURL = authURL
				m.browserVerifier = verifier
				m.browserInput = ""
				m.scherm = schermBrowserLogin
				
				// IPC opzetten
				registerProtocolHandler()
				go openBrowser(authURL)

				// Wacht op de URL via IPC (somtoday://...)
				return m, func() tea.Msg {
					return ipcMsg(<-ipcChan)
				}
			}
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) toetsBrowserLogin(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "esc":
		m.scherm = schermInlogKeuze
	case "enter":
		if m.browserInput != "" {
			m.inlogBezig = true
			schoolUUID := m.gekozenSchool.UUID
			verifier := m.browserVerifier
			input := m.browserInput
			sessie := m.sessie
			return m, func() tea.Msg {
				err := sessie.voltooiBrowserLogin(schoolUUID, verifier, input)
				return inlogMsg{err: err}
			}
		}
	case "backspace":
		if len(m.browserInput) > 0 {
			m.browserInput = m.browserInput[:len(m.browserInput)-1]
		}
	case "ctrl+v": // Pasting usually comes as separate key events or block text, but if handled char by char
		// This is just a fallback. Bubbletea handles pasting automatically if you just listen to runes.
	default:
		if len(key) == 1 {
			r := rune(key[0])
			if unicode.IsPrint(r) {
				m.browserInput += string(r)
			}
		} else if len(key) > 1 && !strings.Contains(key, "+") && !strings.Contains(key, "esc") {
			// Possibly a pasted string from terminal
			for _, r := range key {
				if unicode.IsPrint(r) {
					m.browserInput += string(r)
				}
			}
		}
	}
	return m, nil
}

func (m model) toetsInloggen(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		m.scherm = schermSchoolZoeken
		return m, nil
	case "tab", "down":
		if m.inlogFocus < 2 {
			m.inlogFocus++
		} else {
			m.inlogFocus = 1
		}
	case "shift+tab", "up":
		if m.inlogFocus > 1 {
			m.inlogFocus--
		} else {
			m.inlogFocus = 2
		}
	case "enter":
		if m.inlogVelden[1] != "" && m.inlogVelden[2] != "" && m.gekozenSchool != nil {
			m.inlogBezig = true
			schoolUUID := m.gekozenSchool.UUID
			gebruiker := m.inlogVelden[1]
			wachtwoord := m.inlogVelden[2]
			sessie := m.sessie
			return m, func() tea.Msg {
				err := sessie.inloggen(schoolUUID, gebruiker, wachtwoord)
				return inlogMsg{err: err}
			}
		}
	case "backspace":
		idx := m.inlogFocus
		if len(m.inlogVelden[idx]) > 0 {
			m.inlogVelden[idx] = m.inlogVelden[idx][:len(m.inlogVelden[idx])-1]
		}
	default:
		if len(key) == 1 {
			r := rune(key[0])
			if unicode.IsPrint(r) {
				idx := m.inlogFocus
				m.inlogVelden[idx] += string(r)
			}
		} else if key == "space" {
			idx := m.inlogFocus
			m.inlogVelden[idx] += " "
		}
	}
	return m, nil
}

func (m model) toetsMenu(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.menuSelectie > 0 {
			m.menuSelectie--
		}
	case "down", "j":
		if m.menuSelectie < 5 { // nu 6 opties (0-5) met Profiel
			m.menuSelectie++
		}
	case "u", "U":
		configWissen()
		m.uitloggen = true
		return m, tea.Quit
	case "enter":
		return m.menuActie()
	}
	return m, nil
}

func (m model) toetsDataView(msg tea.KeyMsg) (model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "backspace":
		m.scherm = schermHoofdmenu
		m.scrollOffset = 0
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}
	case "down", "j":
		m.scrollOffset++
	}
	return m, nil
}

func (m model) menuActie() (model, tea.Cmd) {
	m.laden = true
	m.fout = nil
	sessie := m.sessie

	switch m.menuSelectie {
	case 0: // Rooster
		return m, func() tea.Msg {
			vandaag := time.Now().Format("2006-01-02")
			items, err := sessie.roosterOphalen(vandaag)
			return roosterMsg{rooster: items, err: err}
		}
	case 1: // Cijfers
		return m, func() tea.Msg {
			items, err := sessie.cijfersOphalen()
			return cijfersMsg{cijfers: items, err: err}
		}
	case 2: // Huiswerk
		return m, func() tea.Msg {
			items, err := sessie.huiswerkOphalen()
			return huiswerkMsg{huiswerk: items, err: err}
		}
	case 3: // Absenties
		return m, func() tea.Msg {
			items, err := sessie.absentiesOphalen()
			return absentiesMsg{absenties: items, err: err}
		}
	case 4: // Profiel
		m.laden = false
		m.scherm = schermProfiel
		return m, nil
	case 5: // Uitloggen
		configWissen()
		m.sessie = nieuweSessie()
		m.leerling = nil
		m.laden = false
		m.scherm = schermSchoolZoeken
		m.scholen = []School{}
		m.schoolFilter = ""
	}
	return m, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (m model) gefilterdeScholen() []School {
	return m.scholen
}

func (m model) maxBreedte() int {
	w := m.breedte
	if w > 90 {
		w = 90
	}
	if w < 40 {
		w = 40
	}
	return w
}

func (m model) contentHoogte() int {
	return m.hoogte - 6 // header + footer ruimte
}

// ─── View ─────────────────────────────────────────────────────────────────────

func (m model) View() string {
	if m.breedte == 0 || m.hoogte == 0 {
		return ""
	}

	maxW := m.maxBreedte()

	var content string
	switch m.scherm {
	case schermLaden:
		content = m.viewLaden(maxW)
	case schermSchoolZoeken:
		content = m.viewSchoolZoeken(maxW)
	case schermInlogKeuze:
		content = m.viewInlogKeuze(maxW)
	case schermBrowserLogin:
		content = m.viewBrowserLogin(maxW)
	case schermInloggen:
		content = m.viewInloggen(maxW)
	case schermHoofdmenu:
		content = m.viewHoofdmenu(maxW)
	case schermRooster:
		content = m.viewRooster(maxW)
	case schermCijfers:
		content = m.viewCijfers(maxW)
	case schermAbsenties:
		content = m.viewAbsenties(maxW)
	case schermHuiswerk:
		content = m.viewHuiswerk(maxW)
	case schermProfiel:
		content = m.viewProfiel(maxW)
	case schermFout:
		content = m.viewFout(maxW)
	}

	// Centreren in het scherm
	return lipgloss.Place(
		m.breedte, m.hoogte,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

// ─── Schermen ─────────────────────────────────────────────────────────────────

func (m model) viewLaden(w int) string {
	tekst := "⏳ Laden..."
	if m.statusTekst != "" {
		tekst = "⏳ " + m.statusTekst
	}
	return stijlKader.Width(w - 4).Render(
		m.header() + "\n\n" +
			lipgloss.NewStyle().
				Foreground(lipgloss.Color(kleurGroen)).
				Bold(true).
				Render(tekst),
	)
}

func (m model) viewFout(w int) string {
	foutTekst := "Onbekende fout"
	if m.fout != nil {
		foutTekst = m.fout.Error()
	}
	return stijlKader.Width(w - 4).Render(
		m.header() + "\n\n" +
			stijlFout.Render("❌ Fout: "+foutTekst) + "\n\n" +
			stijlSubtiel.Render("[esc] Terug  •  [q] Afsluiten"),
	)
}

func (m model) viewSchoolZoeken(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("🏫 Selecteer je school"))
	b.WriteString("\n\n")

	// Zoekbalk
	zoekInhoud := m.schoolFilter
	if zoekInhoud == "" {
		zoekInhoud = stijlSubtiel.Render("typ om te zoeken...")
	}
	b.WriteString(stijlZoekbalk.Width(w - 8).Render("🔍 " + zoekInhoud))
	b.WriteString("\n")

	// Schoollijst
	gefilterd := m.gefilterdeScholen()
	maxZichtbaar := m.contentHoogte() - 10
	if maxZichtbaar < 3 {
		maxZichtbaar = 3
	}
	if maxZichtbaar > 15 {
		maxZichtbaar = 15
	}

	if len(gefilterd) == 0 {
		if len(m.schoolFilter) >= 3 {
			b.WriteString(stijlSubtiel.Render("  Druk op Enter om te zoeken op Somtoday"))
		} else {
			b.WriteString(stijlSubtiel.Render("  Typ minimaal 3 tekens en druk op Enter"))
		}
	} else {
		start := 0
		if m.schoolSelectie >= maxZichtbaar {
			start = m.schoolSelectie - maxZichtbaar + 1
		}
		eind := start + maxZichtbaar
		if eind > len(gefilterd) {
			eind = len(gefilterd)
		}

		for i := start; i < eind; i++ {
			school := gefilterd[i]
			regel := fmt.Sprintf("%s — %s", school.Naam, school.Plaats)
			if i == m.schoolSelectie {
				b.WriteString(stijlSchoolItemActief.Render("▸ " + regel))
			} else {
				b.WriteString(stijlSchoolItem.Render("  " + regel))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(stijlSubtiel.Render(fmt.Sprintf("  %d resultaten gevonden", len(gefilterd))))
	}

	b.WriteString("\n\n")
	b.WriteString(stijlVoet.Render("[↑↓] Navigeren  •  [enter] Selecteren  •  [esc] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewInlogKeuze(w int) string {
	var b strings.Builder
	b.WriteString(m.header() + "\n\n")
	b.WriteString(stijlTitel.Render("🔒 Inlogmethode kiezen"))
	b.WriteString("\n\n")

	b.WriteString(stijlSubtiel.Render("Gebruikt jouw school een Google of Microsoft account?"))
	b.WriteString("\nKies dan voor 'Browser (O2A)'.\n\n")

	opties := []string{"Gebruikersnaam en Wachtwoord", "Browser (O2A / SSO)"}
	for i, opt := range opties {
		if i == m.inlogKeuze {
			b.WriteString(stijlSchoolItemActief.Render(fmt.Sprintf("▸ %s", opt)) + "\n")
		} else {
			b.WriteString(stijlSchoolItem.Render(fmt.Sprintf("  %s", opt)) + "\n")
		}
	}

	b.WriteString("\n\n" + stijlVoet.Render("[↑↓] Navigeren  •  [enter] Selecteren  •  [esc] Terug"))
	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewBrowserLogin(w int) string {
	var b strings.Builder
	b.WriteString(m.header() + "\n\n")
	b.WriteString(stijlTitel.Render("🌐 Inloggen via Browser"))
	b.WriteString("\n\n")
	
	b.WriteString(stijlSubtiel.Render("1. Je browser is geopend (of open onderstaande link handmatig):\n"))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(kleurSecundair)).Width(w - 8).Render(m.browserURL))
	b.WriteString("\n\n")
	
	b.WriteString(stijlSubtiel.Render("2. Log in met je schoolaccount. De terminal pikt de login\n"))
	b.WriteString(stijlSubtiel.Render("   vervolgens vanzelf weer op zodra je klaar bent!\n\n"))

	if m.inlogBezig {
		b.WriteString("\n\n" + stijlWaarde.Render("⏳ Bezig met inloggen..."))
	}

	b.WriteString("\n\n" + stijlVoet.Render("[esc] Annuleren"))
	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewInloggen(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("🔑 Inloggen bij SomToday"))
	b.WriteString("\n\n")

	if m.inlogBezig {
		b.WriteString(stijlSucces.Render("⏳ Bezig met inloggen..."))
		return stijlKader.Width(w - 4).Render(b.String())
	}

	// School (alleen-lezen)
	b.WriteString(stijlLabel.Render("School"))
	b.WriteString("\n")
	b.WriteString(stijlInputInactief.Width(w - 12).Render("🏫 " + m.inlogVelden[0]))
	b.WriteString("\n")

	// Gebruikersnaam
	b.WriteString(stijlLabel.Render("Gebruikersnaam"))
	b.WriteString("\n")
	gebruikerInhoud := m.inlogVelden[1]
	if gebruikerInhoud == "" && m.inlogFocus != 1 {
		gebruikerInhoud = stijlSubtiel.Render("voer gebruikersnaam in...")
	}
	if m.inlogFocus == 1 {
		gebruikerInhoud += "▎"
	}
	if m.inlogFocus == 1 {
		b.WriteString(stijlInputActief.Width(w - 12).Render("👤 " + gebruikerInhoud))
	} else {
		b.WriteString(stijlInputInactief.Width(w - 12).Render("👤 " + gebruikerInhoud))
	}
	b.WriteString("\n")

	// Wachtwoord
	b.WriteString(stijlLabel.Render("Wachtwoord"))
	b.WriteString("\n")
	wachtwoordDisplay := strings.Repeat("•", len(m.inlogVelden[2]))
	if wachtwoordDisplay == "" && m.inlogFocus != 2 {
		wachtwoordDisplay = stijlSubtiel.Render("voer wachtwoord in...")
	}
	if m.inlogFocus == 2 {
		wachtwoordDisplay += "▎"
	}
	if m.inlogFocus == 2 {
		b.WriteString(stijlInputActief.Width(w - 12).Render("🔒 " + wachtwoordDisplay))
	} else {
		b.WriteString(stijlInputInactief.Width(w - 12).Render("🔒 " + wachtwoordDisplay))
	}

	b.WriteString("\n\n")
	b.WriteString(stijlVoet.Render("[tab] Volgend veld  •  [enter] Inloggen  •  [esc] Terug"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewHoofdmenu(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")

	// Welkom
	if m.leerling != nil {
		b.WriteString(stijlTitel.Render(fmt.Sprintf("👋 Welkom, %s!", m.leerling.Roepnaam)))
		b.WriteString("\n")
	}

	if m.laden {
		b.WriteString(stijlSucces.Render("⏳ Laden..."))
		return stijlKader.Width(w - 4).Render(b.String())
	}

	menuItems := []string{
		"📅  Rooster",
		"📊  Cijfers",
		"📝  Huiswerk",
		"📋  Absenties",
		"👤  Profiel",
		"🚪  Uitloggen",
	}

	b.WriteString(stijlTitel.Render("≡ HOOFDMENU"))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		stijl := stijlMenu
		if i == m.menuSelectie {
			stijl = stijlMenuActief
			b.WriteString(stijl.Render("▶ " + item))
		} else {
			b.WriteString(stijl.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[↑↓] Navigeren • [enter] Selecteren • [u] Uitloggen • [esc/q] Afsluiten"))
	return stijlKader.Width(w).Render(b.String())
}

func (m model) viewRooster(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("📅 Rooster — " + time.Now().Format("02-01-2006")))
	b.WriteString("\n\n")

	if len(m.rooster) == 0 {
		b.WriteString(stijlSubtiel.Render("Geen lessen vandaag."))
	} else {
		// Sorteer op beginlesuur
		items := make([]Afspraak, len(m.rooster))
		copy(items, m.rooster)
		sort.Slice(items, func(i, j int) bool {
			return items[i].BeginLesuur < items[j].BeginLesuur
		})

		maxZichtbaar := m.contentHoogte() - 8
		if maxZichtbaar < 3 {
			maxZichtbaar = 3
		}

		start := m.scrollOffset
		if start > len(items)-1 {
			start = len(items) - 1
			m.scrollOffset = start
		}
		eind := start + maxZichtbaar
		if eind > len(items) {
			eind = len(items)
		}

		for _, a := range items[start:eind] {
			beginTijd := ""
			if len(a.BeginDatumTijd) >= 16 {
				beginTijd = a.BeginDatumTijd[11:16]
			}
			eindTijd := ""
			if len(a.EindDatumTijd) >= 16 {
				eindTijd = a.EindDatumTijd[11:16]
			}

			uurStr := fmt.Sprintf("%de uur", a.BeginLesuur)
			if a.BeginLesuur != a.EindLesuur && a.EindLesuur > 0 {
				uurStr = fmt.Sprintf("%de-%de uur", a.BeginLesuur, a.EindLesuur)
			}

			lokatie := a.Locatie
			if lokatie == "" {
				lokatie = "—"
			}
			docent := a.AdditionalObjects.DocentAfkortingen
			if docent == "" {
				docent = "—"
			}

			status := stijlSubtiel.Render("✓")
			if strings.Contains(strings.ToLower(a.AfspraakStatus), "geannuleerd") {
				status = stijlUitval.Render("✗ Uitval")
			}

			kaart := fmt.Sprintf(
				"%s  %s\n%s  •  %s  •  📍 %s  •  👨‍🏫 %s\n%s",
				stijlLabel.Render("⏰ "+uurStr),
				stijlSubtiel.Render(beginTijd+"–"+eindTijd),
				stijlWaarde.Bold(true).Render(a.Titel),
				status,
				lokatie,
				docent,
				stijlSubtiel.Render(beknopt(a.Omschrijving, 60)),
			)

			if strings.Contains(strings.ToLower(a.AfspraakStatus), "geannuleerd") {
				b.WriteString(stijlKaderActief.Width(w - 10).Render(kaart))
			} else {
				b.WriteString(stijlKader.Width(w - 10).BorderForeground(lipgloss.Color(kleurGroen)).Render(kaart))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[↑↓] Scrollen  •  [esc] Terug  •  [q] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewCijfers(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("📊 Cijfers"))
	b.WriteString("\n\n")

	if len(m.cijfers) == 0 {
		b.WriteString(stijlSubtiel.Render("Geen cijfers gevonden."))
	} else {
		// Tabel header
		headerFmt := fmt.Sprintf("%-20s %-8s %-8s  %s", "Vak", "Cijfer", "Weging", "Omschrijving")
		b.WriteString(stijlTabelHeader.Width(w - 10).Render(headerFmt))
		b.WriteString("\n")

		maxZichtbaar := m.contentHoogte() - 10
		if maxZichtbaar < 3 {
			maxZichtbaar = 3
		}

		start := m.scrollOffset
		if start > len(m.cijfers)-1 {
			start = len(m.cijfers) - 1
			m.scrollOffset = start
		}
		eind := start + maxZichtbaar
		if eind > len(m.cijfers) {
			eind = len(m.cijfers)
		}

		for _, c := range m.cijfers[start:eind] {
			vakNaam := c.AdditionalObjects.Vak.Naam
			if vakNaam == "" {
				vakNaam = c.AdditionalObjects.Vak.Afkorting
			}

			cijferStr := fmt.Sprintf("%.1f", c.Resultaat)
			if c.ResultaatLabel != "" {
				cijferStr = c.ResultaatLabel
			}

			cijferStijl := stijlCijferVoldoende
			if c.Resultaat > 0 && c.Resultaat < 5.5 {
				cijferStijl = stijlCijferOnvoldoende
			} else if c.Resultaat >= 7.5 {
				cijferStijl = stijlCijferZeerVoldoende
			}

			omschr := beknopt(c.Omschrijving, 25)
			rij := fmt.Sprintf("%-20s %s  %-8s  %s",
				beknopt(vakNaam, 20),
				cijferStijl.Render(fmt.Sprintf("%-8s", cijferStr)),
				fmt.Sprintf("%dx", c.Weging),
				stijlSubtiel.Render(omschr),
			)
			b.WriteString(stijlRij.Render(rij))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(stijlSubtiel.Render(fmt.Sprintf("  %d cijfers totaal", len(m.cijfers))))
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[↑↓] Scrollen  •  [esc] Terug  •  [q] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewAbsenties(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("📋 Absenties"))
	b.WriteString("\n\n")

	if len(m.absenties) == 0 {
		b.WriteString(stijlSubtiel.Render("Geen absenties gevonden."))
	} else {
		maxZichtbaar := m.contentHoogte() - 8
		if maxZichtbaar < 3 {
			maxZichtbaar = 3
		}

		start := m.scrollOffset
		if start > len(m.absenties)-1 {
			start = len(m.absenties) - 1
		}
		eind := start + maxZichtbaar
		if eind > len(m.absenties) {
			eind = len(m.absenties)
		}

		for _, a := range m.absenties[start:eind] {
			soort := a.AbsentieReden.AbsentieSoort
			if soort == "" {
				soort = "Onbekend"
			}

			uur := ""
			if a.Afspraak.BeginLesuur > 0 {
				uur = fmt.Sprintf("%de-%de uur", a.Afspraak.BeginLesuur, a.Afspraak.EindLesuur)
			}

			omschr := a.AbsentieReden.Omschrijving
			if omschr == "" {
				omschr = "—"
			}

			rij := fmt.Sprintf(
				"%s  %s  %s\n%s",
				stijlFout.Render(beknopt(soort, 15)),
				stijlLabel.Render(uur),
				stijlWaarde.Render(a.Afspraak.Titel),
				stijlSubtiel.Render(omschr),
			)
			b.WriteString(stijlKader.Width(w - 10).Render(rij))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(stijlSubtiel.Render(fmt.Sprintf("  %d absenties totaal", len(m.absenties))))
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[↑↓] Scrollen  •  [esc] Terug  •  [q] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewHuiswerk(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("📝 Huiswerk"))
	b.WriteString("\n\n")

	if len(m.huiswerk) == 0 {
		b.WriteString(stijlSubtiel.Render("Geen huiswerk gevonden."))
	} else {
		maxZichtbaar := m.contentHoogte() - 8
		if maxZichtbaar < 3 {
			maxZichtbaar = 3
		}

		start := m.scrollOffset
		if start > len(m.huiswerk)-1 {
			start = len(m.huiswerk) - 1
		}
		eind := start + maxZichtbaar
		if eind > len(m.huiswerk) {
			eind = len(m.huiswerk)
		}

		for _, h := range m.huiswerk[start:eind] {
			status := "❌"
			if h.IsAfgerond {
				status = "✅"
			}

			datum := ""
			if len(h.DatumTijd) >= 10 {
				datum = h.DatumTijd[:10]
			}

			vakNaam := h.Studiewijzer.Naam
			if vakNaam == "" {
				vakNaam = h.Lesgroep.Naam
			}

			titel := h.Titel
			if titel == "" {
				titel = h.Onderwerp
			}

			rij := fmt.Sprintf(
				"%s  %s  %s\n%s",
				status,
				stijlLabel.Render(beknopt(vakNaam, 15)),
				stijlSubtiel.Render(datum),
				stijlWaarde.Render(beknopt(titel, 60)),
			)
			b.WriteString(stijlKader.Width(w - 10).Render(rij))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(stijlSubtiel.Render(fmt.Sprintf("  %d items totaal", len(m.huiswerk))))
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[↑↓] Scrollen  •  [esc] Terug  •  [q] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

func (m model) viewProfiel(w int) string {
	var b strings.Builder

	b.WriteString(m.header())
	b.WriteString("\n\n")
	b.WriteString(stijlTitel.Render("👤 Profiel"))
	b.WriteString("\n\n")

	if m.leerling == nil {
		b.WriteString(stijlSubtiel.Render("Geen leerlinggegevens beschikbaar."))
	} else {
		l := m.leerling
		rijen := []struct{ label, waarde string }{
			{"Naam", l.Roepnaam + " " + l.Achternaam},
			{"Leerlingnummer", fmt.Sprintf("%d", l.Leerlingnummer)},
			{"E-mail", nvt(l.Email)},
			{"Telefoon", nvt(l.Mobiel)},
			{"Geboortedatum", nvt(l.Geboortedatum)},
			{"Geslacht", nvt(l.Geslacht)},
		}

		if m.sessie.SchoolNaam != "" {
			rijen = append([]struct{ label, waarde string }{{"School", m.sessie.SchoolNaam}}, rijen...)
		}

		for _, r := range rijen {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				stijlLabel.Width(18).Render(r.label+":"),
				stijlWaarde.Render(r.waarde),
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(stijlVoet.Render("[esc] Terug  •  [q] Afsluiten"))

	return stijlKader.Width(w - 4).Render(b.String())
}

// ─── Header ───────────────────────────────────────────────────────────────────

func (m model) header() string {
	stijlNaam := lipgloss.NewStyle().Foreground(lipgloss.Color(kleurWit)).Bold(true).Padding(0, 1)
	titel := stijlNaam.Render("SOMTODAY")
	info := ""
	if m.leerling != nil {
		info = stijlHeaderInfo.Render(
			fmt.Sprintf("  👤 %s %s", m.leerling.Roepnaam, m.leerling.Achternaam),
		)
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, titel, info)
}

// ─── Hulpfuncties ─────────────────────────────────────────────────────────────

func beknopt(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func nvt(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
