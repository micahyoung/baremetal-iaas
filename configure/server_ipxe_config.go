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
imgfetch --timeout=5000 /{{.StemcellTarGzPath}} /stemcell.tar.gz || goto start

# clobbers dracuts emergency shell to allow dracut-baremetal-agent to be triggered on:
#  * root detection failures - to debug issues with the stemcell
#  * explicit breaks (see rd.break) - to write stemcell, then the agent settings
imgfetch --timeout=5000 /{{.DracutAgentPath}} /bin/dracut-emergency mode=755 || goto start

# root=UUID={{.PartUUID}}  
#  * try to boot from the disk with the UUID of the stemcell
#  * short-circuts if partition already exists
#  * otherwise dracut polls (see rd.retry) until this partition exists
#  * forces a rewrite when stemcell changes
# rd.break=initqueue
#  * cases emergency shell to trigger twice
#  * before partition detection begins
#  * before sysroot pivot begins, after detected partition is mounted
# rd.hostonly=0
#  * from docs: removes all compiled in configuration of the host system the initramfs image was built on
# rd.retry=600
#  * how long dracut should retry the initqueue to configure devices. This should be long enough to write the stemcell
imgargs vmlinuz root=UUID={{.PartUUID}} rd.break=initqueue rd.hostonly=0 rd.retry=600

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
