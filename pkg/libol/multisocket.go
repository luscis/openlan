package libol

import (
	"strings"
	"sync"
)

type MultiSocketServer struct {
	servers []SocketServer
}

func NewMultiSocketServer(servers ...SocketServer) *MultiSocketServer {
	list := make([]SocketServer, 0, len(servers))
	for _, s := range servers {
		if s != nil {
			list = append(list, s)
		}
	}
	return &MultiSocketServer{servers: list}
}

func (m *MultiSocketServer) Listen() error {
	for _, s := range m.servers {
		if err := s.Listen(); err != nil {
			return err
		}
	}
	return nil
}

func (m *MultiSocketServer) Close() {
	for _, s := range m.servers {
		s.Close()
	}
}

func (m *MultiSocketServer) Accept() {
	var wg sync.WaitGroup
	for _, s := range m.servers {
		wg.Add(1)
		go func(v SocketServer) {
			defer wg.Done()
			v.Accept()
		}(s)
	}
	wg.Wait()
}

func (m *MultiSocketServer) ListClient() <-chan SocketClient {
	out := make(chan SocketClient, 64)
	Go(func() {
		for _, s := range m.servers {
			for c := range s.ListClient() {
				if c == nil {
					break
				}
				out <- c
			}
		}
		out <- nil
	})
	return out
}

func (m *MultiSocketServer) UpdateCrypt(block *BlockCrypt) {
	for _, s := range m.servers {
		s.UpdateCrypt(block)
	}
}

func (m *MultiSocketServer) OffClient(client SocketClient) {
	for _, s := range m.servers {
		s.OffClient(client)
	}
}

func (m *MultiSocketServer) TotalClient() int {
	total := 0
	for _, s := range m.servers {
		total += s.TotalClient()
	}
	return total
}

func (m *MultiSocketServer) Loop(call ServerListener) {
	var wg sync.WaitGroup
	for _, s := range m.servers {
		wg.Add(1)
		go func(v SocketServer) {
			defer wg.Done()
			v.Loop(call)
		}(s)
	}
	wg.Wait()
}

func (m *MultiSocketServer) Read(client SocketClient, readAt ReadClient) {
	for _, s := range m.servers {
		s.Read(client, readAt)
	}
}

func (m *MultiSocketServer) String() string {
	values := make([]string, 0, len(m.servers))
	for _, s := range m.servers {
		values = append(values, s.String())
	}
	return strings.Join(values, ",")
}

func (m *MultiSocketServer) Address() string {
	values := make([]string, 0, len(m.servers))
	for _, s := range m.servers {
		values = append(values, s.Address())
	}
	return strings.Join(values, ",")
}

func (m *MultiSocketServer) Statistics() map[string]int64 {
	stat := map[string]int64{}
	for _, s := range m.servers {
		for k, v := range s.Statistics() {
			stat[k] += v
		}
	}
	return stat
}

func (m *MultiSocketServer) SetTimeout(v int64) {
	for _, s := range m.servers {
		s.SetTimeout(v)
	}
}

