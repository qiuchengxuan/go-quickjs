package quickjs

type Config struct {
	MaxStackSize int  // Use 0 to disable maximum stack size check
	ManualFree   bool // Disable runtime and context finalizer and free quickjs manually
}

func DefaultConfig() Config {
	return Config{MaxStackSize: -1}
}
