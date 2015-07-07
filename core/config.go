package core

type Config struct {
	BinaryName string
	Workspace  string
	RunCommand string
	RunArgs    []string
	Assets     map[string][]byte
}
