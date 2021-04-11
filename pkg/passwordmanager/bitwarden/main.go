package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pquerna/otp/totp"
)

// BWItem bitwarden item containing the login information
type BWItem struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Login  BWLogin   `json:"login"`
	Fields []BWField `json:"fields"`
}

// BWLogin bitwarden login credentials
type BWLogin struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Tenant     string
	TOTP       string
	TOTPSecret string  `json:"totp"`
	Uris       []BWUri `json:"uris"`
}

// BWField bitwarden custom fields
type BWField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int32  `json:"type"`
}

// BWUri bitwarden URI associated with the login credentials
type BWUri struct {
	URI string `json:"uri"`
}

func getBWItems(name ...string) []BWItem {

	bw := exec.Command("bw", "list", "items", "--session", os.Getenv("BW_SESSION"))
	bw.Env = os.Environ()

	bwItems := make([]BWItem, 0)

	output, _ := bw.Output()

	if err := json.Unmarshal(output, &bwItems); err != nil {
		log.Fatal(err)
	}
	filteredItems := make([]BWItem, 0)
	for _, item := range bwItems {
		if len(item.Login.Uris) == 0 {
			continue
		}

		for _, pattern := range name {
			if strings.Contains(item.Name, pattern) {

				if len(item.Fields) > 0 {
					for _, field := range item.Fields {
						if strings.HasPrefix(strings.ToLower(field.Name), "tenant") {
							item.Login.Tenant = field.Value
							break
						}
					}
				}

				if strings.Contains(item.Login.Username, "/") {
					parts := strings.SplitN(item.Login.Username, "/", 2)
					if len(parts) == 2 {
						if item.Login.Tenant != "" {
							item.Login.Tenant = parts[0]
						}
						item.Login.Username = parts[1]
					}
				}

				filteredItems = append(filteredItems, item)
				break
			}
		}
	}
	return filteredItems
}

func main() {

	bwItems := getBWItems("c8y", "cumulocity")

	itemTemplate := `{{ .Name | cyan }} {{ if .Login.Uris }} ({{ (index .Login.Uris 0).Uri | red }}) {{end}} ({{ .Login.Tenant | cyan }}/{{ .Login.Username | cyan }})`

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U00002192 " + itemTemplate,
		Inactive: "  " + itemTemplate,
		// Selected: "{{ .ID }}",
		Selected: " ",
		Details: `
--------- Session ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "ID:" | faint }}	{{ .ID }}
{{ "Uri:" | faint }}	{{ (index .Login.Uris 0).Uri }}
{{ "Tenant:" | faint }}	{{ .Login.Tenant }}
{{ "Username:" | faint }}	{{ .Login.Username }}
`,
	}

	searcher := func(input string, index int) bool {
		item := bwItems[index]
		name := strings.Replace(strings.ToLower(item.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select Cumulocity session",
		Items:     bwItems,
		Templates: templates,
		Size:      8,
		Searcher:  searcher,
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	evalStr := ""
	evalStr += "export C8Y_HOST=" + bwItems[i].Login.Uris[0].URI
	evalStr += "\nexport C8Y_TENANT=" + bwItems[i].Login.Tenant
	evalStr += "\nexport C8Y_USERNAME=" + bwItems[i].Login.Username
	evalStr += "\nexport C8Y_PASSWORD=" + bwItems[i].Login.Password

	if bwItems[i].Login.TOTPSecret != "" {
		now := time.Now()
		totpTime := now
		totpPeriod := 30
		totpNextTransition := totpPeriod - now.Second()%30
		totpExpires := now.Add(time.Duration(totpNextTransition) * time.Second)
		if totpNextTransition < 5 {
			totpTime = now.Add(30 * time.Second)
			totpExpires = now.Add(time.Duration(30+totpNextTransition) * time.Second)
		}
		totpCode, err := getTOTPCode(bwItems[i].Login.TOTPSecret, totpTime)

		if err == nil {
			bwItems[i].Login.TOTP = totpCode
			evalStr += "\nexport C8Y_TOTP=" + bwItems[i].Login.TOTP
			evalStr += fmt.Sprintf("\nexport C8Y_TOTP_EXPIRES=%s", totpExpires.Format(time.RFC3339))
		}
	}

	fmt.Printf("%s\n", evalStr)
}

func getTOTPCode(secret string, t time.Time) (string, error) {
	if t.Year() == 0 {
		t = time.Now()
	}
	return totp.GenerateCode(secret, t)
}
