package _switch

import (
	"net"
	"os/exec"
	"strconv"

	"github.com/luscis/openlan/pkg/api"
	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

const (
	UDPBin = "openudp"
)

func GetStateEncap(mode string, sport, dport int) *nl.XfrmStateEncap {
	if mode == "udp" {
		return &nl.XfrmStateEncap{
			Type:            nl.XFRM_ENCAP_ESPINUDP,
			SrcPort:         sport,
			DstPort:         dport,
			OriginalAddress: net.ParseIP("0.0.0.0"),
		}
	}
	return nil
}

type EspWorker struct {
	*WorkerImpl
	proto    nl.Proto
	mode     nl.Mode
	states   []*models.EspState
	policies []*models.EspPolicy
	spec     *co.ESPSpecifies
}

func NewESPWorker(c *co.Network) *EspWorker {
	w := &EspWorker{
		WorkerImpl: NewWorkerApi(c),
		proto:      nl.XFRM_PROTO_ESP,
		mode:       nl.XFRM_MODE_TUNNEL,
	}
	w.spec, _ = c.Specifies.(*co.ESPSpecifies)
	return w
}

type StateParameters struct {
	spi           int
	local, remote net.IP
	auth, crypt   string
}

func (w *EspWorker) newState(args StateParameters) *nl.XfrmState {
	state := &nl.XfrmState{
		Src:   args.local,
		Dst:   args.remote,
		Proto: w.proto,
		Mode:  w.mode,
		Spi:   args.spi,
		Auth: &nl.XfrmStateAlgo{
			Name: "hmac(sha256)",
			Key:  []byte(args.auth),
		},
		Crypt: &nl.XfrmStateAlgo{
			Name: "cbc(aes)",
			Key:  []byte(args.crypt),
		},
	}
	return state
}

type PolicyParameter struct {
	spi           int
	local, remote net.IP
	src, dst      *net.IPNet
	dir           nl.Dir
	pri           int
}

func (w *EspWorker) newPolicy(args PolicyParameter) *nl.XfrmPolicy {
	policy := &nl.XfrmPolicy{
		Src:      args.src,
		Dst:      args.dst,
		Dir:      args.dir,
		Priority: args.pri,
	}
	tmpl := nl.XfrmPolicyTmpl{
		Src:   args.local,
		Dst:   args.remote,
		Proto: w.proto,
		Mode:  w.mode,
		Spi:   args.spi,
	}
	policy.Tmpls = append(policy.Tmpls, tmpl)
	return policy
}

func (w *EspWorker) addState(ms *models.EspState) {
	spi := ms.Spi
	w.out.Info("EspWorker.addState %s", ms.String())
	if st := w.newState(StateParameters{
		spi, ms.Local, ms.Remote, ms.Auth, ms.Crypt,
	}); st != nil {
		st.Encap = GetStateEncap(ms.Encap, co.EspLocalUdp, ms.RemotePort)
		ms.In = st
	} else {
		return
	}
	if st := w.newState(StateParameters{
		spi, ms.Remote, ms.Local, ms.Auth, ms.Crypt,
	}); st != nil {
		st.Encap = GetStateEncap(ms.Encap, ms.RemotePort, co.EspLocalUdp)
		ms.Out = st
	} else {
		return
	}
	w.states = append(w.states, ms)
	cache.EspState.Add(ms)
}

func (w *EspWorker) delState(ms *models.EspState) {
	w.out.Info("EspWorker.delState %s", ms.String())
	cache.EspState.Del(ms.ID())
}

func (w *EspWorker) addPolicy(mp *models.EspPolicy) {
	spi := mp.Spi
	src, err := libol.ParseNet(mp.Source)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s: %s", mp.String(), err)
		return
	}
	dst, err := libol.ParseNet(mp.Dest)
	if err != nil {
		w.out.Error("EspWorker.addPolicy %s: %s", mp.String(), err)
		return
	}
	w.out.Info("EspWorker.addPolicy %s", mp.String())
	if po := w.newPolicy(PolicyParameter{
		spi, mp.Local, mp.Remote, src, dst, nl.XFRM_DIR_OUT, mp.Priority,
	}); po != nil {
		mp.Out = po
	} else {
		return
	}
	if po := w.newPolicy(PolicyParameter{
		spi, mp.Remote, mp.Local, dst, src, nl.XFRM_DIR_FWD, mp.Priority,
	}); po != nil {
		mp.Fwd = po
	} else {
		return
	}
	if po := w.newPolicy(PolicyParameter{
		spi, mp.Remote, mp.Local, dst, src, nl.XFRM_DIR_IN, mp.Priority,
	}); po != nil {
		mp.In = po
	} else {
		return
	}
	w.policies = append(w.policies, mp)
	cache.EspPolicy.Add(mp)
}

func (w *EspWorker) delPolicy(mp *models.EspPolicy) {
	w.out.Info("EspWorker.delPolicy %s", mp.String())
	cache.EspPolicy.Del(mp.ID())
}

func (w *EspWorker) updateXfrm() {
	for _, mem := range w.spec.Members {
		if mem == nil {
			continue
		}
		state := mem.State
		if state.LocalIp == nil || state.RemoteIp == nil {
			continue
		}

		ms := &models.EspState{
			EspState: &schema.EspState{
				Name:       w.spec.Name,
				Spi:        mem.Spi,
				Local:      state.LocalIp,
				Remote:     state.RemoteIp,
				Proto:      uint8(w.proto),
				Mode:       uint8(w.mode),
				Encap:      state.Encap,
				Auth:       state.Auth,
				Crypt:      state.Crypt,
				RemotePort: state.RemotePort,
			},
		}
		w.addState(ms)
		for _, pol := range mem.Policies {
			if pol == nil || pol.Dest == "" {
				continue
			}
			mp := &models.EspPolicy{
				EspPolicy: &schema.EspPolicy{
					Name:     w.spec.Name,
					Spi:      mem.Spi,
					Local:    state.LocalIp,
					Remote:   state.RemoteIp,
					Source:   pol.Source,
					Dest:     pol.Dest,
					Priority: pol.Priority,
				},
			}
			w.addPolicy(mp)
		}
	}
}

func (w *EspWorker) Initialize() {
	w.WorkerImpl.Initialize()
	w.updateXfrm()
}

func (w *EspWorker) AddRoute(device, src, remote string) error {
	link, err := nl.LinkByName(device)
	if link == nil {
		return err
	}
	// add peer routes.
	dst, err := libol.ParseNet(remote)
	if err != nil {
		return libol.NewErr("%s %s.", err, remote)
	}
	gw := libol.ParseAddr(src)
	rte := &nl.Route{
		Dst:       dst,
		Gw:        gw,
		LinkIndex: link.Attrs().Index,
		Priority:  600,
	}
	w.out.Debug("EspWorker.AddRoute: %s", rte)
	if err := nl.RouteReplace(rte); err != nil {
		return libol.NewErr("%s %s.", err, remote)
	}
	return nil
}

func (w *EspWorker) UpDummy(name, addr, peer string) error {
	link, _ := nl.LinkByName(name)
	if link == nil {
		port := &nl.Dummy{
			LinkAttrs: nl.LinkAttrs{
				TxQLen: -1,
				Name:   name,
			},
		}
		if err := nl.LinkAdd(port); err != nil {
			return err
		}
		link, _ = nl.LinkByName(name)
	}
	if err := nl.LinkSetUp(link); err != nil {
		w.out.Error("EspWorker.UpDummy: %s", err)
	}
	if addr != "" {
		ipAddr, err := nl.ParseAddr(addr)
		if err != nil {
			return libol.NewErr("%s %s.", err, addr)
		}
		if err := nl.AddrReplace(link, ipAddr); err != nil {
			w.out.Warn("EspWorker.UpDummy: %s", err)
		}
	}
	w.out.Info("EspWorker.Open %s success", name)
	return nil
}

func (w *EspWorker) addXfrm() {
	for _, state := range w.states {
		w.out.Debug("EspWorker.AddXfrm State %v", state)
		if err := nl.XfrmStateAdd(state.In); err != nil {
			w.out.Error("EspWorker.addXfrm in %s: %s", state.String(), err)
		}
		if err := nl.XfrmStateAdd(state.Out); err != nil {
			w.out.Error("EspWorker.addXfrm out %s: %s", state.String(), err)
		}
	}
	for _, pol := range w.policies {
		w.out.Debug("EspWorker.AddXfrm Policy %v", pol)
		if err := nl.XfrmPolicyAdd(pol.In); err != nil {
			w.out.Error("EspWorker.addXfrm in %v: %s", pol.In, err)
		}
		if err := nl.XfrmPolicyAdd(pol.Fwd); err != nil {
			w.out.Error("EspWorker.addXfrm fwd %v: %s", pol.Fwd, err)
		}
		if err := nl.XfrmPolicyAdd(pol.Out); err != nil {
			w.out.Error("EspWorker.addXfrm out %v: %s", pol.Out, err)
		}
	}
}

func (w *EspWorker) Start(v api.Switcher) {
	w.uuid = v.UUID()
	w.upMember()
	w.addXfrm()
	cache.Esp.Add(&models.Esp{
		Name:    w.cfg.Name,
		Address: w.spec.Address,
	})
	w.WorkerImpl.Start(v)
}

func (w *EspWorker) DownDummy(name string) error {
	link, _ := nl.LinkByName(name)
	if link == nil {
		return nil
	}
	port := &nl.Dummy{
		LinkAttrs: nl.LinkAttrs{
			TxQLen: -1,
			Name:   name,
		},
	}
	if err := nl.LinkDel(port); err != nil {
		return err
	}
	return nil
}

func (w *EspWorker) delXfrm() {
	for _, mp := range w.policies {
		w.delPolicy(mp)
		if err := nl.XfrmPolicyDel(mp.In); err != nil {
			w.out.Warn("EspWorker.delXfrm Policy.in %s: %s", mp.String(), err)
		}
		if err := nl.XfrmPolicyDel(mp.Fwd); err != nil {
			w.out.Warn("EspWorker.delXfrm Policy.fwd %s: %s", mp.String(), err)
		}
		if err := nl.XfrmPolicyDel(mp.Out); err != nil {
			w.out.Warn("EspWorker.delXfrm Policy.out %s: %s", mp.String(), err)
		}
	}
	w.policies = nil
	for _, ms := range w.states {
		w.delState(ms)
		if err := nl.XfrmStateDel(ms.In); err != nil {
			w.out.Warn("EspWorker.delXfrm State.in %s: %s", ms.String(), err)
		}
		if err := nl.XfrmStateDel(ms.Out); err != nil {
			w.out.Warn("EspWorker.delXfrm State.out %s: %s", ms.String(), err)
		}
	}
	w.states = nil
}

func (w *EspWorker) Stop() {
	w.WorkerImpl.Stop()
	cache.Esp.Del(w.cfg.Name)
	w.downMember()
	w.delXfrm()
}

func (w *EspWorker) Reload(v api.Switcher) {
	w.Stop()
	w.Initialize()
	w.Start(v)
}

func (w *EspWorker) upMember() {
	for _, mem := range w.spec.Members {
		if mem.Peer == "" {
			continue
		}
		if err := w.UpDummy(w.spec.Name, mem.Address, mem.Peer); err != nil {
			w.out.Warn("EspWorker.UpDummy %d %s", mem.Spi, err)
		}
		for _, po := range mem.Policies {
			if err := w.AddRoute(w.spec.Name, mem.Address, po.Dest); err != nil {
				w.out.Warn("EspWorker.AddRoute %d %s", mem.Spi, err)
			}
		}
	}
}

func (w *EspWorker) downMember() {
	for _, mem := range w.spec.Members {
		if mem.Peer == "" {
			continue
		}
		if err := w.DownDummy(w.spec.Name); err != nil {
			w.out.Error("EspWorker.downMember %d %s", mem.Spi, err)
		}
	}
}

func OpenUDP() {
	libol.Go(func() {
		args := []string{
			"-port", strconv.Itoa(co.EspLocalUdp),
			"-log:file", "/var/openlan/openudp.log",
		}
		libol.Info("%s %v", UDPBin, args)
		cmd := exec.Command(UDPBin, args...)
		if err := cmd.Run(); err != nil {
			libol.Error("OpenUDP %s", err)
		}
	})
}
