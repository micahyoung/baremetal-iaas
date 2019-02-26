package settings

var defaultKernelPath = `vmlinuz`
var defaultInitRDPath = `initrd.img`
var defaultStemcellTarGzPathPath = `stemcell.tar.gz`
var defaultInitBinPath = `init.bin`

type Settings struct {
	KernelPath        string
	InitRDPath        string
	StemcellTarGzPath string
	InitBinPath       string
}

func NewSettings() *Settings {
	return &Settings{defaultKernelPath, defaultInitRDPath, defaultStemcellTarGzPathPath, defaultInitBinPath}
}
