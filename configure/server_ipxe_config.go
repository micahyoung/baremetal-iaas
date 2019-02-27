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

# clobbers dracuts cmdline ask script allow dracut-baremetal-agent:
#  * allows explicit break (see rd.break) to make it run before other stages
imgfetch --timeout=5000 /{{.DracutAgentPath}} /bin/dracut-cmdline-ask mode=755 || goto start

# root=/dev/sda1  
#  * try to boot from the disk (dracut will wait until read)
# rd.cmdline=ask
#  * trigger dracut-agent very early on
# rw
#  * make /sysroot read-write so agent-bootstrap-env.json can be written
imgargs vmlinuz root=/dev/sda1 rd.cmdline=ask rw

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
