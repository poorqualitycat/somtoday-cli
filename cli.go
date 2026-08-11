package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

func runCLI(args []string) {
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	command := args[0]
	sessie := nieuweSessie()

	switch command {
	case "help":
		printHelp()
		os.Exit(0)
	case "login":
		handleCLILogin(sessie, args[1:])
	case "logout":
		configWissen()
		fmt.Println("Succesvol uitgelogd.")
		os.Exit(0)
	default:
		if !sessie.isIngelogd() {
			fmt.Println("Fout: Je bent nog niet ingelogd. Gebruik 'somtoday-cli login' of 'somtoday-tui' om in te loggen.")
			os.Exit(1)
		}
		if err := sessie.tokenVernieuwen(); err != nil {
			fmt.Printf("Sessie verlopen, opnieuw inloggen vereist: %v\n", err)
			os.Exit(1)
		}

		switch command {
		case "cijfers":
			handleCLICijfers(sessie, args[1:])
		case "planning":
			handleCLIPlanning(sessie, args[1:])
		case "huiswerk":
			handleCLIHuiswerk(sessie, args[1:])
		case "afwezigheid":
			handleCLIAfwezigheid(sessie, args[1:])
		case "school-info":
			handleCLISchoolInfo(sessie, args[1:])
		case "berichten":
			handleCLIBerichten(sessie, args[1:])
		default:
			fmt.Printf("Onbekend commando: %s\n", command)
			printHelp()
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Println(`Alle somtoday-cli commands

*somtoday-cli login
    --school "school_naam" (hoeft niet exact te zijn rond af op bovenste zoekresultaat)
    --sso OF --username "user" --password "pswd"
    --help
    (login als er al config bestaat)

somtoday-cli cijfers
    --vak "vaknaam"
    --sort "hoog/laag" / "laag/hoog"
    --help

somtoday-cli planning
   --week "weeknummer" 
   --jaar "jaartal"
   --dag "dag van week"
   --tijd "tijd"
   --eerstvolgende 
   --huidige

somtoday-cli huiswerk
   [planning parameters]
   --voltooid 
   --onvoltooid
   --beschrijving "vak"
   --voltooi "vak"
   --onvoltooi "vak"

somtoday-cli afwezigheid
somtoday-cli school-info
somtoday-cli berichten
somtoday-cli logout`)
}

func handleCLILogin(sessie *Sessie, args []string) {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	school := fs.String("school", "", "Naam van de school")
	sso := fs.Bool("sso", false, "Gebruik SSO login")
	username := fs.String("username", "", "Gebruikersnaam")
	password := fs.String("password", "", "Wachtwoord")
	fs.Parse(args)

	if sessie.isIngelogd() {
		fmt.Println("Je bent al ingelogd.")
		os.Exit(0)
	}

	if *school == "" {
		fmt.Println("Fout: --school is verplicht.")
		os.Exit(1)
	}

	// Zoek school
	scholen, err := searchOrganisaties(*school)
	if err != nil || len(scholen) == 0 {
		fmt.Println("Fout: School niet gevonden.")
		os.Exit(1)
	}
	gekozen := scholen[0]
	fmt.Printf("School gevonden: %s\n", gekozen.Naam)

	if *sso {
		if *username != "" || *password != "" {
			fmt.Println("Fout: --sso kan niet samen met --username of --password.")
			os.Exit(1)
		}
		// SSO Flow
		authURL, verifier, err := startBrowserLogin(gekozen.UUID)
		if err != nil {
			fmt.Printf("Fout bij voorbereiden SSO: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Open de volgende URL in je browser om in te loggen:")
		fmt.Println(authURL)
		
		registerProtocolHandler()
		go openBrowser(authURL)
		
		// Wacht op de URL via IPC
		redirectURL := <-ipcChan
		if err := sessie.voltooiBrowserLogin(gekozen.UUID, verifier, redirectURL); err != nil {
			fmt.Printf("SSO inloggen mislukt: %v\n", err)
			os.Exit(1)
		}
		sessie.SchoolNaam = gekozen.Naam
		configSet(keySchoolNaam, gekozen.Naam)
		fmt.Println("Succesvol ingelogd via SSO!")

	} else {
		if *username == "" || *password == "" {
			fmt.Println("Fout: Geef --sso OF (--username EN --password).")
			os.Exit(1)
		}
		err := sessie.inloggen(gekozen.UUID, *username, *password)
		if err != nil {
			fmt.Printf("Fout bij inloggen: %v\n", err)
			os.Exit(1)
		}
		sessie.SchoolNaam = gekozen.Naam
		configSet(keySchoolNaam, gekozen.Naam)
		fmt.Println("Succesvol ingelogd!")
	}
}

func handleCLICijfers(sessie *Sessie, args []string) {
	fs := flag.NewFlagSet("cijfers", flag.ExitOnError)
	vakFilter := fs.String("vak", "", "Filter op vaknaam")
	sortFilter := fs.String("sort", "", "Sorteer op hoog/laag of laag/hoog")
	fs.Parse(args)

	cijfers, err := sessie.cijfersOphalen()
	if err != nil {
		fmt.Printf("Fout bij ophalen cijfers: %v\n", err)
		os.Exit(1)
	}

	var gefilterd []Cijfer
	for _, c := range cijfers {
		vakNaam := c.AdditionalObjects.Vak.Naam
		if vakNaam == "" {
			vakNaam = c.AdditionalObjects.Vak.Afkorting
		}
		if *vakFilter != "" && !strings.Contains(strings.ToLower(vakNaam), strings.ToLower(*vakFilter)) {
			continue
		}
		gefilterd = append(gefilterd, c)
	}

	if *sortFilter == "hoog/laag" {
		sort.Slice(gefilterd, func(i, j int) bool { return gefilterd[i].Resultaat > gefilterd[j].Resultaat })
	} else if *sortFilter == "laag/hoog" {
		sort.Slice(gefilterd, func(i, j int) bool { return gefilterd[i].Resultaat < gefilterd[j].Resultaat })
	}

	if len(gefilterd) == 0 {
		fmt.Println("Geen cijfers gevonden.")
		return
	}

	for _, c := range gefilterd {
		vakNaam := c.AdditionalObjects.Vak.Naam
		if vakNaam == "" {
			vakNaam = c.AdditionalObjects.Vak.Afkorting
		}
		cijferStr := fmt.Sprintf("%.1f", c.Resultaat)
		if c.ResultaatLabel != "" {
			cijferStr = c.ResultaatLabel
		}
		fmt.Printf("- %s: %s (Datum: %s)\n", vakNaam, cijferStr, c.DatumInvoer)
	}
}

func handleCLIPlanning(sessie *Sessie, args []string) {
	fs := flag.NewFlagSet("planning", flag.ExitOnError)
	week := fs.String("week", "", "Weeknummer")
	jaar := fs.String("jaar", "", "Jaartal")
	_ = week
	_ = jaar
	dag := fs.String("dag", "", "Dag van de week (bijv. maandag, 1, ma)")
	tijd := fs.String("tijd", "", "Tijd (bijv. 8:30)")
	eerstvolgende := fs.Bool("eerstvolgende", false, "Toon eerstvolgende vak")
	huidige := fs.Bool("huidige", false, "Toon huidige vak")
	fs.Parse(args)

	// Datum logica
	datum := time.Now()
	// Eenvoudige datum afhandeling, kan later met iso/week uitgebreid worden
	// Voor nu fallback naar vandaag
	if *dag != "" {
		// Complexere logica voor dag nodig afhankelijk van week/jaar
	}
	
	datumStr := datum.Format("2006-01-02")
	afspraken, err := sessie.roosterOphalen(datumStr)
	if err != nil {
		fmt.Printf("Fout bij ophalen planning: %v\n", err)
		os.Exit(1)
	}

	if *huidige {
		// Logica voor huidige vak
		if len(afspraken) > 0 {
			a := afspraken[0]
			fmt.Printf("Huidig vak: %de uur: %s (%s)\n", a.BeginLesuur, a.Titel, a.Omschrijving)
		}
		return
	}

	if *eerstvolgende {
		// Logica voor eerstvolgende vak
		if len(afspraken) > 0 {
			a := afspraken[0]
			fmt.Printf("Eerstvolgende vak: %de uur: %s (%s)\n", a.BeginLesuur, a.Titel, a.Omschrijving)
		}
		return
	}

	if *tijd != "" {
		// Zoek specifieke tijd
		fmt.Printf("Vak rond tijdstip %s...\n", *tijd)
		return
	}

	if len(afspraken) == 0 {
		fmt.Println("Geen planning voor deze dag.")
		return
	}

	for _, a := range afspraken {
		fmt.Printf("- %de uur: %s (%s)\n", a.BeginLesuur, a.Titel, a.Omschrijving)
	}
}

func handleCLIHuiswerk(sessie *Sessie, args []string) {
	fs := flag.NewFlagSet("huiswerk", flag.ExitOnError)
	voltooid := fs.Bool("voltooid", false, "Toont alleen voltooide taken")
	onvoltooid := fs.Bool("onvoltooid", false, "Toont alleen onvoltooide taken")
	beschrijving := fs.String("beschrijving", "", "Toon beschrijving voor vak")
	fs.Parse(args)

	hw, err := sessie.huiswerkOphalen()
	if err != nil {
		fmt.Printf("Fout bij ophalen huiswerk: %v\n", err)
		os.Exit(1)
	}

	if len(hw) == 0 {
		fmt.Println("Geen huiswerk gevonden.")
		return
	}

	for _, h := range hw {
		if *voltooid && !h.IsAfgerond { continue }
		if *onvoltooid && h.IsAfgerond { continue }
		
		status := "[ ]"
		if h.IsAfgerond {
			status = "[X]"
		}
		vak := h.Lesgroep.Naam
		if vak == "" {
			vak = h.Studiewijzer.Naam
		}
		if *beschrijving != "" && !strings.Contains(strings.ToLower(vak), strings.ToLower(*beschrijving)) {
			continue
		}
		fmt.Printf("%s %s: %s\n", status, vak, h.Omschrijving)
	}
}

func handleCLIAfwezigheid(sessie *Sessie, args []string) {
	ab, err := sessie.absentiesOphalen()
	if err != nil {
		fmt.Printf("Fout bij ophalen absenties: %v\n", err)
		os.Exit(1)
	}

	if len(ab) == 0 {
		fmt.Println("Geen afwezigheden gevonden voor het huidige schooljaar.")
		return
	}

	for _, a := range ab {
		fmt.Printf("- %s: %s\n", a.BeginDatumTijd, a.AbsentieReden.Omschrijving)
	}
}

func handleCLISchoolInfo(sessie *Sessie, args []string) {
	fmt.Printf("Ingelogd bij: %s\n", sessie.SchoolNaam)
}

func handleCLIBerichten(sessie *Sessie, args []string) {
	fmt.Println("Berichten ophalen (nog in aanbouw).")
}
