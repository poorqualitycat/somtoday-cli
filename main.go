package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const ipcPort = 41337

// ─── Constanten ───────────────────────────────────────────────────────────────

const (
	keyringService = "SomToday-CLI"
	keyBaseURL     = "base_url"
	keyAccess      = "access_token"
	keyRefresh     = "refresh_token"
	keyStudentID   = "student_id"
	keySchoolUUID  = "school_uuid"
	keySchoolNaam  = "school_naam"
	keyUsername    = "username"
	keyPassword    = "password"

	// Correcte SomToday client_id — publieke client, GEEN client_secret
	somtodayClientID = "D50E0C06-32D1-4B41-A137-A9A850C892C2"
	somtodayTokenURL = "https://somtoday.nl/oauth2/token"
	somtodayLeerlingClientID = "somtoday-leerling-native"
	somtodayRedirectURI      = "somtoday://nl.topicus.somtoday.leerling/oauth/callback"
)

// ─── OAuth Token Response ─────────────────────────────────────────────────────

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	APIURL       string `json:"somtoday_api_url"`
	Tenant       string `json:"somtoday_tenant"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	// Foutgevallen
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// ─── Scholenlijst ─────────────────────────────────────────────────────────────

type School struct {
	UUID   string `json:"uuid"`
	Naam   string `json:"naam"`
	Plaats string `json:"plaats"`
}

type organisatieGroep struct {
	Instellingen []School `json:"instellingen"`
}

// ─── Leerling ─────────────────────────────────────────────────────────────────

type Leerling struct {
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Leerlingnummer int    `json:"leerlingnummer"`
	Roepnaam       string `json:"roepnaam"`
	Achternaam     string `json:"achternaam"`
	Email          string `json:"email"`
	Mobiel         string `json:"mobiel"`
	Geboortedatum  string `json:"geboortedatum"`
	Geslacht       string `json:"geslacht"`
}

type LeerlingenResponse struct {
	Items []Leerling `json:"items"`
}

// ─── Cijfers ──────────────────────────────────────────────────────────────────

type Cijfer struct {
	AdditionalObjects struct {
		Vak struct {
			Afkorting string `json:"afkorting"`
			Naam      string `json:"naam"`
		} `json:"vak"`
	} `json:"additionalObjects"`
	Resultaat              float64 `json:"resultaat"`
	Weging                 int     `json:"weging"`
	Omschrijving           string  `json:"omschrijving"`
	DatumInvoer            string  `json:"datumInvoer"`
	ResultaatLabel         string  `json:"resultaatLabel"`
	ResultaatLabelAfkorting string `json:"resultaatLabelAfkorting"`
	IsExamencijfer         bool    `json:"isExamencijfer"`
}

type CijfersResponse struct {
	Items []Cijfer `json:"items"`
}

// ─── Rooster ──────────────────────────────────────────────────────────────────

type Afspraak struct {
	AdditionalObjects struct {
		DocentAfkortingen string `json:"docentAfkortingen"`
	} `json:"additionalObjects"`
	AfspraakType struct {
		Naam         string `json:"naam"`
		Omschrijving string `json:"omschrijving"`
	} `json:"afspraakType"`
	Locatie        string `json:"locatie"`
	BeginDatumTijd string `json:"beginDatumTijd"`
	EindDatumTijd  string `json:"eindDatumTijd"`
	Titel          string `json:"titel"`
	Omschrijving   string `json:"omschrijving"`
	AfspraakStatus string `json:"afspraakStatus"`
	BeginLesuur    int    `json:"beginLesuur"`
	EindLesuur     int    `json:"eindLesuur"`
}

type RoosterResponse struct {
	Items []Afspraak `json:"items"`
}

// ─── Absenties ────────────────────────────────────────────────────────────────

type Absentie struct {
	Afspraak struct {
		BeginDatumTijd string `json:"beginDatumTijd"`
		EindDatumTijd  string `json:"eindDatumTijd"`
		Titel          string `json:"titel"`
		BeginLesuur    int    `json:"beginLesuur"`
		EindLesuur     int    `json:"eindLesuur"`
	} `json:"afspraak"`
	AbsentieReden struct {
		AbsentieSoort string `json:"absentieSoort"`
		Omschrijving  string `json:"omschrijving"`
	} `json:"absentieReden"`
	BeginDatumTijd string `json:"beginDatumTijd"`
}

type AbsentiesResponse struct {
	Items []Absentie `json:"items"`
}

// ─── Huiswerk ─────────────────────────────────────────────────────────────────

type Huiswerk struct {
	Studiewijzer struct {
		Naam string `json:"naam"`
	} `json:"studiewijzer"`
	DatumTijd    string `json:"datumTijd"`
	Lesgroep     struct {
		Naam string `json:"naam"`
	} `json:"lesgroep"`
	Titel        string `json:"titel"`
	Omschrijving string `json:"omschrijving"`
	Onderwerp    string `json:"onderwerp"`
	IsAfgerond   bool   `json:"isAfgerond"`
}

type HuiswerkResponse struct {
	Items []Huiswerk `json:"items"`
}

// ─── Sessie (in-memory state) ─────────────────────────────────────────────────

type Sessie struct {
	AccessToken  string
	RefreshToken string
	BaseURL      string
	StudentID    string
	SchoolUUID   string
	SchoolNaam   string
	Client       *http.Client
}

func nieuweSessie() *Sessie {
	s := &Sessie{
		BaseURL: "https://api.somtoday.nl",
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	
	// Laad configuratie als deze bestaat
	if val, _ := configGet(keyAccess); val != "" {
		s.AccessToken = val
	}
	if val, _ := configGet(keyRefresh); val != "" {
		s.RefreshToken = val
	}
	if val, _ := configGet(keyBaseURL); val != "" {
		s.BaseURL = val
	}
	if val, _ := configGet(keyStudentID); val != "" {
		s.StudentID = val
	}
	if val, _ := configGet(keySchoolUUID); val != "" {
		s.SchoolUUID = val
	}
	if val, _ := configGet(keySchoolNaam); val != "" {
		s.SchoolNaam = val
	}
	
	return s
}

func (s *Sessie) isIngelogd() bool {
	return s.AccessToken != ""
}

// ─── Scholen ophalen ──────────────────────────────────────────────────────────

func searchOrganisaties(term string) ([]School, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	req1, err := http.NewRequest("GET", "https://inloggen.somtoday.nl/", nil)
	if err != nil {
		return nil, err
	}
	req1.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp1, err := client.Do(req1)
	if err != nil {
		return nil, err
	}
	body1, _ := io.ReadAll(resp1.Body)
	resp1.Body.Close()

	re := regexp.MustCompile(`source: '\./(\?[^']+)'`)
	matches := re.FindStringSubmatch(string(body1))
	if len(matches) < 2 {
		return nil, fmt.Errorf("kon zoek-endpoint niet vinden in Somtoday portaal")
	}
	ajaxPath := matches[1]

	ajaxURL := "https://inloggen.somtoday.nl/" + ajaxPath + "&term=" + url.QueryEscape(term)
	req2, err := http.NewRequest("GET", ajaxURL, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Referer", "https://inloggen.somtoday.nl/")
	req2.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp2, err := client.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	if resp2.StatusCode != 200 {
		return nil, fmt.Errorf("fout bij zoeken scholen (status %d)", resp2.StatusCode)
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(body2, &data); err != nil {
		return nil, fmt.Errorf("ongeldig antwoord van Somtoday: %v", err)
	}

	var scholen []School
	for _, s := range data {
		id, _ := s["id"].(string)
		val, _ := s["value"].(string)
		label, _ := s["label"].(string)

		plaats := ""
		if parts := strings.Split(label, " - "); len(parts) > 1 {
			plaats = parts[len(parts)-1]
		}
		
		scholen = append(scholen, School{
			UUID:   id,
			Naam:   val,
			Plaats: plaats,
		})
	}
	return scholen, nil
}

// ─── OAuth: Inloggen (password grant) ─────────────────────────────────────────

func (s *Sessie) inloggen(schoolUUID, gebruikersnaam, wachtwoord string) error {
	values := url.Values{}
	values.Set("grant_type", "password")
	values.Set("username", schoolUUID+"\\"+gebruikersnaam)
	values.Set("password", wachtwoord)
	values.Set("scope", "openid")
	values.Set("client_id", somtodayClientID)

	req, err := http.NewRequest("POST", somtodayTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("kon verzoek niet aanmaken: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("inlogverzoek mislukt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("kon response niet lezen: %w", err)
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return fmt.Errorf("kon token niet parsen: %w", err)
	}

	if token.Error != "" {
		return fmt.Errorf("inloggen mislukt: %s (%s)", token.Error, token.ErrorDescription)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("inloggen mislukt (status %d)", resp.StatusCode)
	}

	s.AccessToken = token.AccessToken
	s.RefreshToken = token.RefreshToken
	s.SchoolUUID = schoolUUID
	if token.APIURL != "" {
		s.BaseURL = strings.TrimRight(token.APIURL, "/")
	}

	// Opslaan in configuratie
	configSet(keyAccess, s.AccessToken)
	configSet(keyRefresh, s.RefreshToken)
	configSet(keyBaseURL, s.BaseURL)
	configSet(keySchoolUUID, s.SchoolUUID)
	configSet(keyUsername, gebruikersnaam)
	configSet(keyPassword, wachtwoord)

	return nil
}

// ─── OAuth: Token vernieuwen ──────────────────────────────────────────────────

func (s *Sessie) tokenVernieuwen() error {
	if s.RefreshToken == "" {
		return errors.New("geen refresh token beschikbaar")
	}

	values := url.Values{}
	values.Set("grant_type", "refresh_token")
	values.Set("refresh_token", s.RefreshToken)
	values.Set("client_id", somtodayClientID)
	values.Set("scope", "openid")

	req, err := http.NewRequest("POST", somtodayTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("kon verzoek niet aanmaken: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("token vernieuwen mislukt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return fmt.Errorf("kon token niet parsen: %w", err)
	}

	if token.Error != "" {
		return fmt.Errorf("vernieuwen mislukt: %s", token.ErrorDescription)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("vernieuwen mislukt (status %d)", resp.StatusCode)
	}

	s.AccessToken = token.AccessToken
	if token.RefreshToken != "" {
		s.RefreshToken = token.RefreshToken
	}
	if token.APIURL != "" {
		s.BaseURL = strings.TrimRight(token.APIURL, "/")
	}

	configSet(keyAccess, s.AccessToken)
	configSet(keyRefresh, s.RefreshToken)
	configSet(keyBaseURL, s.BaseURL)

	return nil
}

// ─── OAuth: O2A (Browser) Inloggen ────────────────────────────────────────────

func genereerPKCE() (verifier, challenge string) {
	b := make([]byte, 32)
	rand.Read(b)
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func genereerState() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func startBrowserLogin(schoolUUID string) (authURL string, verifier string, err error) {
	verifier, challenge := genereerPKCE()
	state := genereerState()

	u, _ := url.Parse("https://inloggen.somtoday.nl/oauth2/authorize")
	q := u.Query()
	q.Set("redirect_uri", somtodayRedirectURI)
	q.Set("client_id", somtodayLeerlingClientID)
	q.Set("state", state)
	q.Set("response_type", "code")
	q.Set("scope", "openid")
	q.Set("tenant_uuid", schoolUUID)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()

	return u.String(), verifier, nil
}


func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func registerProtocolHandler() {
	if runtime.GOOS == "linux" {
		exe, err := os.Executable()
		if err != nil {
			return
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return
		}
		desktopDir := filepath.Join(home, ".local", "share", "applications")
		os.MkdirAll(desktopDir, 0755)
		desktopFile := filepath.Join(desktopDir, "somtoday-cli.desktop")
		content := fmt.Sprintf(`[Desktop Entry]
Name=Somtoday CLI
Exec=%s %%u
Type=Application
Terminal=true
MimeType=x-scheme-handler/somtoday;
`, exe)
		os.WriteFile(desktopFile, []byte(content), 0644)
		exec.Command("xdg-mime", "default", "somtoday-cli.desktop", "x-scheme-handler/somtoday").Run()
	}
}

var ipcChan = make(chan string)

func startIPCListener() {
	go func() {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ipcPort))
		if err != nil {
			return
		}
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			buf := make([]byte, 2048)
			n, err := conn.Read(buf)
			if err == nil {
				ipcChan <- strings.TrimSpace(string(buf[:n]))
			}
			conn.Close()
		}
	}()
}

func (s *Sessie) voltooiBrowserLogin(schoolUUID, verifier, callbackURL string) error {
	u, err := url.Parse(callbackURL)
	if err != nil {
		return fmt.Errorf("ongeldige callback URL: %w", err)
	}
	code := u.Query().Get("code")
	if code == "" {
		return errors.New("geen 'code' gevonden in de URL. Is het inloggen gelukt?")
	}

	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", somtodayRedirectURI)
	values.Set("code_verifier", verifier)
	values.Set("code", code)
	values.Set("scope", "openid")
	values.Set("client_id", somtodayLeerlingClientID)

	req, err := http.NewRequest("POST", somtodayTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("kon verzoek niet aanmaken: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("inlogverzoek mislukt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("kon response niet lezen: %w", err)
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return fmt.Errorf("kon token niet parsen: %w", err)
	}

	if token.Error != "" {
		return fmt.Errorf("inloggen mislukt: %s (%s)", token.Error, token.ErrorDescription)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("inloggen mislukt (status %d)", resp.StatusCode)
	}

	s.AccessToken = token.AccessToken
	s.RefreshToken = token.RefreshToken
	s.SchoolUUID = schoolUUID

	if token.APIURL != "" {
		s.BaseURL = strings.TrimRight(token.APIURL, "/")
	}

	// Sla op in keyring
	configSet(keyAccess, s.AccessToken)
	configSet(keyRefresh, s.RefreshToken)
	configSet(keySchoolUUID, s.SchoolUUID)
	configSet(keyBaseURL, s.BaseURL)
	return nil
}

// ─── API verzoeken ────────────────────────────────────────────────────────────

func (s *Sessie) apiGet(endpoint string) ([]byte, error) {
	reqURL := s.BaseURL + "/rest/v1" + endpoint
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API verzoek mislukt: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 401 {
		// Probeer token te vernieuwen en opnieuw te proberen
		if refreshErr := s.tokenVernieuwen(); refreshErr != nil {
			return nil, fmt.Errorf("niet geautoriseerd en vernieuwen mislukt: %w", refreshErr)
		}
		return s.apiGet(endpoint)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("API fout (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ─── Data ophalen ─────────────────────────────────────────────────────────────

func (s *Sessie) leerlingOphalen() (*Leerling, error) {
	data, err := s.apiGet("/leerlingen")
	if err != nil {
		return nil, err
	}
	var resp LeerlingenResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, errors.New("geen leerling gevonden")
	}

	// Student ID opslaan
	leerling := &resp.Items[0]
	if len(leerling.Links) > 0 {
		parts := strings.Split(leerling.Links[0].Href, "/")
		s.StudentID = parts[len(parts)-1]
		configSet(keyStudentID, s.StudentID)
	}

	return leerling, nil
}

func (s *Sessie) cijfersOphalen() ([]Cijfer, error) {
	if s.StudentID == "" {
		if _, err := s.leerlingOphalen(); err != nil {
			return nil, err
		}
	}

	endpoints := []string{
		"/resultaten/huidigVoorLeerling/" + s.StudentID,
		"/resultaten/huidig",
		"/resultaten?leerling.id=" + s.StudentID,
		"/leerlingen/" + s.StudentID + "/resultaten/huidig",
		"/leerlingen/" + s.StudentID + "/resultaten",
	}

	var data []byte
	var err error
	var resp CijfersResponse

	for _, ep := range endpoints {
		data, err = s.apiGet(ep)
		if err == nil {
			if jsonErr := json.Unmarshal(data, &resp); jsonErr == nil {
				return resp.Items, nil // Kan leeg zijn, dat is prima
			}
		} else if strings.Contains(err.Error(), "404") {
			// Somtoday retourneert vaak 404 als er simpelweg (nog) geen resultaten zijn in het nieuwe schooljaar
			continue
		}
	}

	if err != nil && strings.Contains(err.Error(), "404") {
		return []Cijfer{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("geen cijfers kunnen ophalen, laatste API fout: %v", err)
	}
	return []Cijfer{}, nil
}

func (s *Sessie) roosterOphalen(datum string) ([]Afspraak, error) {
	endpoint := fmt.Sprintf("/afspraken?begindatum=%s&einddatum=%s&additional=vpilesopnames&additional=docentAfkortingen&additional=lpilesopnames", datum, datum)
	data, err := s.apiGet(endpoint)
	if err != nil {
		return nil, err
	}
	var resp RoosterResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (s *Sessie) absentiesOphalen() ([]Absentie, error) {
	data, err := s.apiGet("/absentiemeldingen")
	if err != nil {
		return nil, err
	}
	var resp AbsentiesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	// Filter voor huidig schooljaar (vanaf 1 augustus)
	now := time.Now()
	startJaar := now.Year()
	if now.Month() < time.August {
		startJaar--
	}
	startDatumStr := fmt.Sprintf("%04d-08-01", startJaar)

	var gefilterd []Absentie
	for _, a := range resp.Items {
		dateStr := a.BeginDatumTijd
		if dateStr == "" && a.Afspraak.BeginDatumTijd != "" {
			dateStr = a.Afspraak.BeginDatumTijd
		}
		if dateStr != "" && len(dateStr) >= 10 {
			if dateStr[:10] < startDatumStr {
				continue
			}
		}
		gefilterd = append(gefilterd, a)
	}

	return gefilterd, nil
}

func (s *Sessie) huiswerkOphalen() ([]Huiswerk, error) {
	data, err := s.apiGet("/studiewijzeritemdagtoekenningen")
	if err != nil {
		return nil, err
	}
	var resp HuiswerkResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// ─── Config helpers ────────────────────────────────────────────────────────────

func getConfigPath() (string, error) {
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = filepath.Join(os.Getenv("APPDATA"), "somtoday-cli")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, "Library", "Preferences", "somtoday-cli")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config", "somtoday-cli")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func configGet(key string) (string, error) {
	path, err := getConfigPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil // Bestand bestaat nog niet
	}
	var cfg map[string]string
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", nil
	}
	return strings.TrimSpace(cfg[key]), nil
}

func configSet(key, value string) {
	if value == "" {
		return
	}
	path, err := getConfigPath()
	if err != nil {
		return
	}
	var cfg map[string]string
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	if cfg == nil {
		cfg = make(map[string]string)
	}
	cfg[key] = value
	data, err = json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		os.WriteFile(path, data, 0600)
	}
}

func configWissen() {
	path, err := getConfigPath()
	if err == nil {
		os.Remove(path)
	}
}

	// ─── Entrypoint ───────────────────────────────────────────────────────────────

func main() {
	// ─── IPC Client: Als we door de browser worden gestart ──────────────
	if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "somtoday://") {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", ipcPort))
		if err == nil {
			conn.Write([]byte(os.Args[1] + "\n"))
			conn.Close()
		}
		os.Exit(0)
	}
	// ──────────────────────────────────────────────────────────────────

	exeName := filepath.Base(os.Args[0])
	isTUI := strings.Contains(strings.ToLower(exeName), "tui") || (len(os.Args) == 1 && !strings.Contains(strings.ToLower(exeName), "cli"))

	if isTUI {
		startIPCListener()
		for {
			m := startTUI()
			if !m.uitloggen {
				break
			}
		}
	} else {
		runCLI(os.Args[1:])
	}
}
