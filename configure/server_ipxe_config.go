package configure

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/micahyoung/baremetal-iaas/settings"
)

var serverScriptTemplate = strings.TrimSpace(`
#!ipxe

:start
dhcp || goto start

imgfree

kernel --timeout=5000 /{{.KernelPath}} || goto start
initrd --timeout=5000 /{{.InitRDPath}} || goto start

#TODO imgfetch --timeout=5000 /agent-bootstrap-env.json /agent-bootstrap-env.json || goto start
#DEBUG imgfetch --timeout=30000 /{{.StemcellTarGzPath}} /stemcell.tar.gz || goto start
imgfetch --timeout=7000 /{{.StemcellTarGzPath}} /stemcell.tar.gz || goto start

imgfetch --timeout=5000 /{{.InitBinPath}} /bin/dracut-emergency mode=755 || goto start

# Try to boot the BaremetalIaas partition
# otherwise make dracut fail fast which will user the replacement dracut-emergency 
imgargs vmlinuz root=UUID={{.PartUUID}} rd.hostonly=0 rd.auto=0 rd.retry=0
boot
`)

type serverScriptData struct {
	*settings.Settings
	PartUUID string
}

func (g *Generator) ServerScriptContent(settings *settings.Settings, partUUID string) (string, error) {
	var err error

	data := &serverScriptData{settings, partUUID}

	var content bytes.Buffer
	err = template.Must(template.New("server").Parse(serverScriptTemplate)).Execute(&content, data)
	if err != nil {
		return "", err
	}

	return content.String(), nil
}
