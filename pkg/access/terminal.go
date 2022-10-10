package access

import (
	"fmt"
	"github.com/chzyer/readline"
	"github.com/luscis/openlan/pkg/libol"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
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
	cur := os.Getenv("PWD")
	home := os.Getenv("HOME")
	if strings.HasPrefix(cur, home) {
		cur = strings.Replace(cur, home, "~", 1)
	}
	return fmt.Sprintf("[%s:%s]> ", user, cur)
}

func (t *Terminal) CmdEdit(args []string) {
}

func (t *Terminal) CmdShow(args []string) {
	action := ""
	if len(args) > 1 {
		action = args[1]
	}
	switch action {
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

func (t *Terminal) CmdBye(args []string) {
}

func (t *Terminal) CmdMode(args []string) {
	if len(args) < 2 {
		return
	}
	switch args[1] {
	case "vi":
		t.Console.SetVimMode(true)
	case "emacs":
		t.Console.SetVimMode(false)
	}
}

func (t *Terminal) CmdShell(args []string) {
	cmd := args[0]
	var c2 *exec.Cmd
	if _, err := exec.LookPath(cmd); err == nil {
		c2 = exec.Command(cmd, args[1:]...)
	} else {
		params := append([]string{
			"-c",
		}, args...)
		c2 = exec.Command("/bin/bash", params...)
	}
	c2.Stdin = os.Stdin
	c2.Stdout = os.Stdout
	c2.Stderr = os.Stderr
	if err := c2.Run(); err != nil {
		fmt.Println(err)
	}
}

func (t *Terminal) CmdHelp(args []string) {
	fmt.Printf("Usage: COMMAND ARGS...\n")
	fmt.Printf("  show\t Display configuration\n")
	fmt.Printf("  quit\t Exit\n")
	fmt.Printf("  edit\t Edit configuration\n")
}

func (t *Terminal) signal() {
	x := make(chan os.Signal)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
}

func (t *Terminal) loop() {
	for {
		line, err := t.Console.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			break
		}
		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}
		cmd := args[0]
		switch {
		case cmd == "":
			break
		case cmd == "?":
			t.CmdHelp(args)
		case cmd == "mode":
			t.CmdMode(args)
		case cmd == "show":
			t.CmdShow(args)
		case cmd == "edit":
			t.CmdEdit(args)
		case cmd == "exit" || cmd == "quit":
			t.CmdBye(args)
			goto quit
		default:
			t.CmdShell(args)
		}
	}
quit:
	fmt.Printf("Terminal.Start quit")
}

func (t *Terminal) Start() {
	if t.Console == nil {
		return
	}
	defer t.Console.Close()

	t.signal()
	t.loop()
}
