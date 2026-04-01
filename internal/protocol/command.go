package protocol

type Name string

const (
	CmdPing   Name = "PING"
	CmdSet    Name = "SET"
	CmdGet    Name = "GET"
	CmdDel    Name = "DEL"
	CmdExists Name = "EXISTS"
)

type Command struct {
	Name Name
	Args []string
	Raw  string
}

func (c Command) IsWrite() bool {
	return c.Name == CmdSet || c.Name == CmdDel
}
