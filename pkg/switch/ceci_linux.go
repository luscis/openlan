package cswitch

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/luscis/openlan/pkg/api"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

const (
	CeciBin = "/usr/bin/openceci"
	CeciDir = "/var/openlan/ceci/"
)

type CeciWorker struct {
	*WorkerImpl
	spec *co.CeciSpecifies
}

func NewCeciWorker(c *co.Network) *CeciWorker {
	w := &CeciWorker{
		WorkerImpl: NewWorkerApi(c),
	}
	api.Call.SetCeciApi(w)
	w.spec, _ = c.Specifies.(*co.CeciSpecifies)
	return w
}

func (w *CeciWorker) Initialize() {
	w.out.Info("CeciWorker.Initialize")
	w.addCache()
}

func (w *CeciWorker) killPid(name string) {
	if libol.FileExist(name) != nil {
		return
	}

	pid, _ := os.ReadFile(name)
	kill, _ := exec.LookPath("kill")
	cmd := exec.Command(kill, strings.TrimSpace(string(pid)))
	if err := cmd.Run(); err != nil {
		w.out.Warn("CeciWorker.killPid:%s: %s", pid, err)
		return
	}
}

func (w *CeciWorker) findPid(name string) int {
	pid := 0
	data, err := os.ReadFile(name)
	if err != nil {
		return 0
	}
	if n, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
		pid = n
	}
	return pid
}

func (w *CeciWorker) isRunning(name string) bool {
	pid := w.findPid(name)
	return pid > 0 && libol.HasProcess(pid)
}

func (w *CeciWorker) processStatusByPidFile(file string) string {
	if w.findPid(file) <= 0 {
		return "stopped"
	}
	if w.isRunning(file) {
		return "running"
	}
	return "stopped"
}

func (w *CeciWorker) loadCeciStats(listen string) *schema.CeciStats {
	file := CeciDir + listen + ".stats"
	stats := &schema.CeciStats{}
	if err := libol.UnmarshalLoad(stats, file); err != nil {
		return nil
	}
	if stats.StartAt == "" && stats.Total == 0 && stats.Bytes == 0 {
		return nil
	}
	return stats
}

func (w *CeciWorker) start(obj *co.CeciProxy) {
	name := CeciDir + obj.Id()
	runtimeCert, err := w.prepareRuntimeCert(name, obj.Cert)
	if err != nil {
		w.out.Warn("CeciWorker.start: %s", err)
		return
	}

	var data any
	switch obj.Mode {
	case "http":
		data = &co.HttpProxy{
			Listen:   obj.Listen,
			Network:  obj.Network,
			Backends: obj.Backends,
			Cert:     runtimeCert,
		}
	case "name":
		nameTo := ""
		if len(obj.Target) > 0 {
			nameTo = obj.Target[0]
		}
		data = &co.NameProxy{
			Listen:   obj.Listen,
			Nameto:   nameTo,
			Backends: obj.Backends,
		}
	default:
		data = &co.TcpProxy{
			Listen: obj.Listen,
			Target: obj.Target,
		}
	}
	if err := libol.MarshalSave(data, name+".yaml", true); err != nil {
		w.out.Warn("CeciWorker.start: %s", err)
		return
	}

	libol.Go(func() {
		w.out.Info("CeciWorker.start: %s", obj.Id())
		args := []string{"-mode", obj.Mode, "-conf", name + ".yaml"}
		if obj.Network != "" {
			args = append(args, "-network", obj.Network)
		}
		if obj.Mode == "http" {
			args = append(args, "-stats:file", name+".stats")
		}
		args = append(args, "-log:file", name+".log")
		args = append(args, "-write-pid", name+".pid")
		cmd := exec.Command(CeciBin, args...)
		if err := cmd.Run(); err != nil {
			w.out.Warn("CeciWorker.start: %s %s", args, err)
			return
		}
	})
}

func normalizeBalance(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "leastconn", "source", "first", "uri", "url_param", "hdr":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "roundrobin"
	}
}

func normalizeServiceProtocol(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "http":
		return "http"
	default:
		return "tcp"
	}
}

func (w *CeciWorker) startService(obj *co.CeciService) {
	name := CeciDir + obj.Id()
	protocol := normalizeServiceProtocol(obj.Protocol)
	balance := normalizeBalance(obj.Balance)
	type routeItem struct {
		servers []string
		match   []string
	}
	routes := make([]routeItem, 0, len(obj.Routes)+len(obj.Backends))
	for _, r := range obj.Routes {
		match := make([]string, 0, len(r.Match))
		for _, host := range r.Match {
			host = strings.TrimSpace(host)
			if host != "" {
				match = append(match, host)
			}
		}
		servers := make([]string, 0, len(r.Backends))
		for _, backend := range r.Backends {
			backend = strings.TrimSpace(backend)
			if backend != "" {
				servers = append(servers, backend)
			}
		}
		if len(servers) == 0 {
			continue
		}
		routes = append(routes, routeItem{servers: servers, match: match})
	}
	for _, server := range obj.Backends {
		server = strings.TrimSpace(server)
		if server != "" {
			routes = append(routes, routeItem{servers: []string{server}})
		}
	}
	if len(routes) == 0 {
		w.out.Warn("CeciWorker.startService: no backend configured")
		return
	}

	cfg := []string{
		"global",
		"  log /dev/log local0",
		"  log-tag openlan-ceci-service",
		"  maxconn 4096",
		"",
		"defaults",
		"  log global",
		"  option tcplog",
		"  mode " + protocol,
		"  timeout connect 5s",
		"  timeout client 1m",
		"  timeout server 1m",
		"",
		"frontend openlan_service_frontend",
		"  bind " + obj.Listen,
	}
	if protocol == "http" {
		cfg = append(cfg, "  option httplog")
		for i, r := range routes {
			if len(r.match) == 0 {
				continue
			}
			acl := fmt.Sprintf("host_route_%d", i+1)
			cfg = append(cfg, fmt.Sprintf("  acl %s hdr(host),lower,field(1,:) -i %s", acl, strings.Join(r.match, " ")))
			cfg = append(cfg, fmt.Sprintf("  use_backend openlan_service_backend_%d if %s", i+1, acl))
		}
		cfg = append(cfg, "  default_backend openlan_service_backend_default", "")
		hasDefault := false
		for _, r := range routes {
			if len(r.match) == 0 {
				hasDefault = true
				break
			}
		}
		for i, r := range routes {
			cfg = append(cfg, fmt.Sprintf("backend openlan_service_backend_%d", i+1))
			cfg = append(cfg, "  mode http")
			cfg = append(cfg, "  balance "+balance)
			for j, server := range r.servers {
				cfg = append(cfg, fmt.Sprintf("  server srv%d_%d %s check", i+1, j+1, server))
			}
			cfg = append(cfg, "")
		}
		cfg = append(cfg, "backend openlan_service_backend_default")
		cfg = append(cfg, "  mode http")
		cfg = append(cfg, "  balance "+balance)
		n := 0
		for _, r := range routes {
			if len(r.match) > 0 && hasDefault {
				continue
			}
			for _, server := range r.servers {
				n++
				cfg = append(cfg, fmt.Sprintf("  server srv%d %s check", n, server))
			}
		}
	} else {
		cfg = append(cfg, "  default_backend openlan_service_backend", "", "backend openlan_service_backend", "  balance "+balance)
		for i, r := range routes {
			for j, server := range r.servers {
				cfg = append(cfg, fmt.Sprintf("  server srv%d_%d %s check", i+1, j+1, server))
			}
		}
	}
	confFile := name + ".haproxy.cfg"
	if err := os.WriteFile(confFile, []byte(strings.Join(cfg, "\n")+"\n"), 0600); err != nil {
		w.out.Warn("CeciWorker.startService: write cfg: %s", err)
		return
	}

	haproxy, err := exec.LookPath("haproxy")
	if err != nil {
		w.out.Warn("CeciWorker.startService: haproxy not found: %s", err)
		return
	}
	pidFile := name + ".pid"

	libol.Go(func() {
		cmd := exec.Command(haproxy, "-f", confFile, "-db")
		if err := cmd.Start(); err != nil {
			w.out.Warn("CeciWorker.startService: start: %s", err)
			return
		}
		_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", cmd.Process.Pid)), 0644)
		w.out.Info("CeciWorker.startService: %s", obj.Id())
		if err := cmd.Wait(); err != nil {
			w.out.Warn("CeciWorker.startService: wait: %s", err)
		}
	})
}

func (w *CeciWorker) restart(obj *co.CeciProxy) {
	name := CeciDir + obj.Id()
	w.killPid(name + ".pid")
	w.start(obj)
}

func (w *CeciWorker) restartService(obj *co.CeciService) {
	name := CeciDir + obj.Id()
	w.killPid(name + ".pid")
	w.startService(obj)
}

func (w *CeciWorker) saveCertFile(name string, data string) (string, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return "", nil
	}
	if err := os.WriteFile(name, []byte(data+"\n"), 0600); err != nil {
		return "", err
	}
	return name, nil
}

func normalizeCeciCert(data *schema.Cert) (*co.Cert, error) {
	if data == nil {
		return nil, nil
	}

	cert := &co.Cert{
		CrtFile:  data.CrtFile,
		KeyFile:  data.KeyFile,
		CaFile:   data.CaFile,
		CrtData:  data.CrtData,
		KeyData:  data.KeyData,
		CaData:   data.CaData,
		Insecure: data.Insecure,
	}
	if err := cert.LoadData(); err != nil {
		return nil, err
	}
	if cert.CrtData == "" || cert.KeyData == "" {
		return nil, libol.NewErr("certificate and key are required")
	}
	return cert, nil
}

func (w *CeciWorker) prepareRuntimeCert(name string, cert *co.Cert) (*co.Cert, error) {
	if cert == nil {
		return nil, nil
	}
	current := *cert
	if err := current.LoadData(); err != nil {
		return nil, err
	}
	if current.CrtData == "" || current.KeyData == "" {
		return nil, libol.NewErr("certificate content is required")
	}
	crtFile, err := w.saveCertFile(name+".crt", current.CrtData)
	if err != nil {
		return nil, err
	}
	keyFile, err := w.saveCertFile(name+".key", current.KeyData)
	if err != nil {
		return nil, err
	}
	runtime := &co.Cert{
		CrtFile:  crtFile,
		KeyFile:  keyFile,
		Insecure: current.Insecure,
	}
	if current.CaData != "" {
		caFile, err := w.saveCertFile(name+".ca.crt", current.CaData)
		if err != nil {
			return nil, err
		}
		runtime.CaFile = caFile
	}
	return runtime, nil
}

func (w *CeciWorker) Start(v api.SwitchApi) {
	w.uuid = v.UUID()
	w.out.Info("CeciWorker.Start")

	if w.spec == nil {
		return
	}
	w.spec.ListProxy(func(obj *co.CeciProxy) {
		if obj == nil {
			return
		}
		if w.isRunning(CeciDir + obj.Id() + ".pid") {
			w.out.Info("CeciWorker.Start: already running %s", obj.Id())
			return
		}
		w.start(obj)
	})
	w.spec.ListService(func(obj *co.CeciService) {
		if obj == nil {
			return
		}
		if w.isRunning(CeciDir + obj.Id() + ".pid") {
			w.out.Info("CeciWorker.Start: already running %s", obj.Id())
			return
		}
		w.startService(obj)
	})
}

func (w *CeciWorker) Stop(kill bool) {
	w.out.Info("CeciWorker.Stop")
	if !kill {
		return
	}
	if w.spec == nil {
		return
	}
	w.spec.ListProxy(func(obj *co.CeciProxy) {
		if obj == nil {
			return
		}
		name := CeciDir + obj.Id()
		if w.isRunning(name + ".pid") {
			w.killPid(name + ".pid")
		}
	})
	w.spec.ListService(func(obj *co.CeciService) {
		if obj == nil {
			return
		}
		name := CeciDir + obj.Id()
		if w.isRunning(name + ".pid") {
			w.killPid(name + ".pid")
		}
	})
}

func (w *CeciWorker) AddProxy(data schema.CeciProxy) error {
	obj := &co.CeciProxy{
		Mode:    data.Mode,
		Listen:  data.Listen,
		Network: data.Network,
		Target:  data.Target,
	}
	if len(data.Backends) > 0 {
		obj.Backends = make(co.ToForwards, 0, len(data.Backends))
		for _, backend := range data.Backends {
			obj.Backends = append(obj.Backends, &co.ForwardTo{
				Match:    backend.Match,
				Protocol: backend.Protocol,
				Insecure: backend.Insecure,
				Secret:   backend.Secret,
				Nameto:   backend.Nameto,
			})
		}
	}
	if cert, err := normalizeCeciCert(data.Cert); err != nil {
		w.out.Warn("CeciWorker.AddProxy: %s", err)
		return err
	} else {
		obj.Cert = cert
	}

	obj.Correct()
	if t, _ := w.spec.FindProxy(obj); t == nil {
		w.spec.AddProxy(obj)
	} else {
		t.Mode = data.Mode
		t.Network = data.Network
		t.Target = data.Target
		t.Backends = obj.Backends
		t.Cert = obj.Cert
	}

	w.restart(obj)
	return nil
}

func (w *CeciWorker) RestartProxy(data schema.CeciProxy) error {
	obj := &co.CeciProxy{
		Listen: data.Listen,
	}
	obj.Correct()
	if current, _ := w.spec.FindProxy(obj); current != nil {
		w.restart(current)
		return nil
	}
	return libol.NewErr("ceci entry not found")
}

func (w *CeciWorker) DelProxy(data schema.CeciProxy) {
	obj := &co.CeciProxy{
		Listen: data.Listen,
	}
	obj.Correct()
	if _, removed := w.spec.DelProxy(obj); removed {
		name := CeciDir + obj.Id()
		w.killPid(name + ".pid")
	}
}

func (w *CeciWorker) AddService(data schema.CeciProxy) error {
	service := &co.CeciService{
		Listen:  data.Listen,
		Network: data.Network,
	}
	if data.Service != nil {
		service.Protocol = data.Service.Protocol
		service.Balance = data.Service.Balance
		if len(data.Service.Backends) > 0 || len(data.Service.Routes) > 0 {
			return libol.NewErr("add service does not support backends/routes, please use service backend add")
		}
	}
	if len(data.Target) > 0 {
		return libol.NewErr("add service does not support target, please use service backend add")
	}
	service.Correct()
	if t, _ := w.spec.FindService(service); t == nil {
		w.spec.AddService(service)
	} else {
		t.Network = service.Network
		t.Protocol = service.Protocol
		t.Balance = service.Balance
		t.Backends = service.Backends
		t.Routes = service.Routes
	}
	w.restartService(service)
	return nil
}

func (w *CeciWorker) RestartService(data schema.CeciProxy) error {
	obj := &co.CeciService{Listen: data.Listen}
	obj.Correct()
	if current, _ := w.spec.FindService(obj); current != nil {
		w.restartService(current)
		return nil
	}
	return libol.NewErr("ceci service not found")
}

func (w *CeciWorker) DelService(data schema.CeciProxy) {
	obj := &co.CeciService{Listen: data.Listen}
	obj.Correct()
	if _, removed := w.spec.DelService(obj); removed {
		name := CeciDir + obj.Id()
		w.killPid(name + ".pid")
	}
}

func (w *CeciWorker) AddServiceBackend(data schema.CeciServiceBackendAdd) error {
	obj := &co.CeciService{Listen: data.Listen}
	obj.Correct()
	current, _ := w.spec.FindService(obj)
	if current == nil {
		return libol.NewErr("ceci service not found")
	}
	if !current.AddBackend(data.Hostname, data.Backends) {
		return libol.NewErr("backend is required")
	}
	w.restartService(current)
	return nil
}

func (w *CeciWorker) ListProxy(call func(obj schema.CeciProxy)) {
	if call == nil {
		return
	}
	if w.spec == nil {
		return
	}
	w.spec.ListProxy(func(value *co.CeciProxy) {
		if value == nil {
			return
		}
		item := schema.CeciProxy{
			Mode:    value.Mode,
			Listen:  value.Listen,
			Network: value.Network,
			Target:  value.Target,
			Cert: func() *schema.Cert {
				if value.Cert == nil {
					return nil
				}
				out := &schema.Cert{
					Insecure: value.Cert.Insecure,
					CrtData:  value.Cert.CrtData,
					KeyData:  value.Cert.KeyData,
					CaData:   value.Cert.CaData,
				}
				if out.CrtData == "" && value.Cert.CrtFile != "" {
					if data, err := os.ReadFile(value.Cert.CrtFile); err == nil {
						out.CrtData = string(data)
					}
				}
				if out.KeyData == "" && value.Cert.KeyFile != "" {
					if data, err := os.ReadFile(value.Cert.KeyFile); err == nil {
						out.KeyData = string(data)
					}
				}
				if out.CaData == "" && value.Cert.CaFile != "" {
					if data, err := os.ReadFile(value.Cert.CaFile); err == nil {
						out.CaData = string(data)
					}
				}
				return out
			}(),
			Backends: func() []schema.ForwardTo {
				out := make([]schema.ForwardTo, 0, len(value.Backends))
				for _, backend := range value.Backends {
					if backend == nil {
						continue
					}
					out = append(out, schema.ForwardTo{
						Server:   backend.Server,
						Match:    backend.Match,
						Protocol: backend.Protocol,
						Insecure: backend.Insecure,
						Secret:   backend.Secret,
						Nameto:   backend.Nameto,
					})
				}
				return out
			}(),
			Stats:  w.loadCeciStats(value.Listen),
			Status: w.processStatusByPidFile(CeciDir + value.Id() + ".pid"),
		}
		call(item)
	})
}

func (w *CeciWorker) ListService(call func(obj schema.CeciProxy)) {
	if call == nil {
		return
	}
	if w.spec == nil {
		return
	}
	w.spec.ListService(func(value *co.CeciService) {
		if value == nil {
			return
		}
		item := schema.CeciProxy{
			Mode:    "service",
			Listen:  value.Listen,
			Network: value.Network,
			Service: &schema.CeciService{
				Protocol: value.Protocol,
				Balance:  value.Balance,
				Backends: value.Backends,
				Routes: func() []schema.CeciServiceBackend {
					out := make([]schema.CeciServiceBackend, 0, len(value.Routes))
					for _, route := range value.Routes {
						if len(route.Backends) == 0 {
							continue
						}
						out = append(out, schema.CeciServiceBackend{
							Backends: route.Backends,
							Match:    route.Match,
						})
					}
					return out
				}(),
			},
			Stats:  w.loadCeciStats(value.Listen),
			Status: w.processStatusByPidFile(CeciDir + value.Id() + ".pid"),
		}
		call(item)
	})
}
