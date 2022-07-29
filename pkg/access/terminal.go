package access

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/luscis/openlan/pkg/libol"
	"io"
	"strings"
)

type Terminal struct {
	Pointer Pointer
	Console *readline.Instance
}

func NewTerminal(pointer Pointer) *Terminal {
	t := &Terminal{Pointer: pointer}
	completer := readline.NewPrefixCompleter(
		readline.PcItem("quit"),
		readline.PcItem("help"),
		readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("show",
			readline.PcItem("config"),
			readline.PcItem("network"),
			readline.PcItem("record"),
			readline.PcItem("statistics"),
		),
		readline.PcItem("edit",
			readline.PcItem("user"),
			readline.PcItem("connection"),
		),
	)

	config := &readline.Config{
		Prompt:            t.Prompt(),
		HistoryFile:       ".history",
		InterruptPrompt:   "^C",
		EOFPrompt:         "quit",
		HistorySearchFold: true,
		AutoComplete:      completer,
	}
	if l, err := readline.NewEx(config); err == nil {
		t.Console = l
	}
	return t
}

func (t *Terminal) Prompt() string {
	user := t.Pointer.User()
	alias := t.Pointer.Alias()
	tenant := t.Pointer.Tenant()
	return fmt.Sprintf("[%s@%s %s]# ", user, alias, tenant)
}

func (t *Terminal) CmdEdit(args string) {
}

func (t *Terminal) CmdShow(args string) {
	switch args {
	case "record":
		v := t.Pointer.Record()
		if out, err := libol.Marshal(v, true); err == nil {
			fmt.Printf("%s\n", out)
		}
	case "statistics":
		if c := t.Pointer.Client(); c != nil {
			v := c.Statistics()
			if out, err := libol.Marshal(v, true); err == nil {
				fmt.Printf("%s\n", out)
			}
		}
	case "config":
		cfg := t.Pointer.Config()
		if str, err := libol.Marshal(cfg, true); err == nil {
			fmt.Printf("%s\n", str)
		}
	case "network":
		cfg := t.Pointer.Network()
		if str, err := libol.Marshal(cfg, true); err == nil {
			fmt.Printf("%s\n", str)
		}
	default:
		v := struct {
			UUID   string
			UpTime int64
			Device string
			Status string
		}{
			UUID:   t.Pointer.UUID(),
			UpTime: t.Pointer.UpTime(),
			Device: t.Pointer.IfName(),
			Status: t.Pointer.Status().String(),
		}
		if str, err := libol.Marshal(v, true); err == nil {
			fmt.Printf("%s\n", str)
		}
	}
}

func (t *Terminal) Trim(v string) string {
	return strings.TrimSpace(v)
}

func (t *Terminal) CmdBye() {
}

func (t *Terminal) CmdMode(args string) {
	switch args {
	case "vi":
		t.Console.SetVimMode(true)
	case "emacs":
		t.Console.SetVimMode(false)
	}
}

func (t *Terminal) Start() {
	if t.Console == nil {
		return
	}
	defer t.Console.Close()
	for {
		line, err := t.Console.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = t.Trim(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			t.CmdMode(t.Trim(line[5:]))
		case line == "show":
			t.CmdShow("")
		case line == "quit" || line == "exit":
			t.CmdBye()
			goto quit
		case strings.HasPrefix(line, "show "):
			t.CmdShow(t.Trim(line[5:]))
		case strings.HasPrefix(line, "edit "):
			t.CmdEdit(t.Trim(line[5:]))
		}
	}
quit:
	fmt.Printf("Terminal.Start quit")
}
