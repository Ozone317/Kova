package core

type KovaCmd struct {
	Cmd  string
	Args []string
}

type KovaCmds []*KovaCmd
