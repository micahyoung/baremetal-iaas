package settings

var defaultKernelPath = `vmlinuz`
var defaultInitRDPath = `initrd.img`
var defaultStemcellTarGzPathPath = `stemcell.tar.gz`
var defaultDracutAgentPath = `dracut-agent`

type Settings struct {
	KernelPath        string
	InitRDPath        string
	StemcellTarGzPath string
	DracutAgentPath   string
}

func NewSettings() *Settings {
	return &Settings{defaultKernelPath, defaultInitRDPath, defaultStemcellTarGzPathPath, defaultDracutAgentPath}
}
