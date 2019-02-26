package configure

import (
	"bytes"
	"strings"
	"text/template"
)

var embeddedScriptTemplate = strings.TrimSpace(`
#!ipxe

:start

# obtain IP or reboot to retry
dhcp || goto start
route

# boot from server or reboot to retry
chain --timeout=5000 {{.ServerBaseURL}}/{{.ServerScriptName}} || goto start
`)

type embeddedScriptData struct {
	ServerBaseURL    string
	ServerScriptName string
}

func (g *Generator) EmbeddedScriptContent(serverBaseURL, serverScriptName string) (string, error) {
	var err error

	data := &embeddedScriptData{serverBaseURL, serverScriptName}

	var content bytes.Buffer
	err = template.Must(template.New("embedded").Parse(embeddedScriptTemplate)).Execute(&content, data)
	if err != nil {
		return "", err
	}
	return content.String(), nil
}
