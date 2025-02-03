package quickjs

type GlobalConfig struct {
	ManualFree bool
}

var globalConfig GlobalConfig

// Disable runtime and context finalizer and free quickjs manually
func SetManualFree() {
	globalConfig.ManualFree = true
}

type Config struct {
	MaxStackSize int // Use 0 to disable maximum stack size check
}

func DefaultConfig() Config {
	return Config{MaxStackSize: -1}
}
