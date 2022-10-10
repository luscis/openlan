package app

import (
	"encoding/json"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	"strings"
)

type Request struct {
	master Master
}

func NewRequest(m Master) *Request {
	return &Request{
		master: m,
	}
}

func (r *Request) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	out := client.Out()
	if frame.IsEthernet() {
		return nil
	}
	if out.Has(libol.DEBUG) {
		out.Log("Request.OnFrame %s.", frame)
	}
	action, body := frame.CmdAndParams()
	if out.Has(libol.CMD) {
		out.Cmd("Request.OnFrame: %s %s", action, body)
	}
	switch action {
	case libol.NeighborReq:
		r.onNeighbor(client, body)
	case libol.IpAddrReq:
		r.onIpAddr(client, body)
	case libol.LeftReq:
		r.onLeave(client, body)
	case libol.LoginReq:
		out.Debug("Request.OnFrame %s: %s", action, body)
	default:
		r.onDefault(client, body)
	}
	return nil
}

func (r *Request) onDefault(client libol.SocketClient, data []byte) {
	m := libol.NewControlFrame(libol.PongResp, data)
	_ = client.WriteMsg(m)
}

func (r *Request) onNeighbor(client libol.SocketClient, data []byte) {
	resp := make([]schema.Neighbor, 0, 32)
	for obj := range cache.Neighbor.List() {
		if obj == nil {
			break
		}
		resp = append(resp, models.NewNeighborSchema(obj))
	}
	if respStr, err := json.Marshal(resp); err == nil {
		m := libol.NewControlFrame(libol.NeighborResp, respStr)
		_ = client.WriteMsg(m)
	}
}

func (r *Request) findLease(ifAddr string, p *models.Point, n *models.Network) *schema.Lease {
	if n == nil {
		return nil
	}
	alias := p.Alias
	network := n.Name
	lease := cache.Network.GetLease(alias, network) // try by alias firstly
	if ifAddr == "" {
		if lease == nil { // now to alloc it.
			lease = cache.Network.NewLease(alias, network)
		}
	} else {
		ipAddr := strings.SplitN(ifAddr, "/", 2)[0]
		if lease == nil || lease.Address != ipAddr {
			lease = cache.Network.AddLease(alias, ipAddr, network)
		}
	}
	if lease != nil {
		lease.Client = p.Client.String()
	}
	return lease
}
func (r *Request) onIpAddr(client libol.SocketClient, data []byte) {
	out := client.Out()
	out.Info("Request.onIpAddr: %s", data)
	recv := models.NewNetwork("", "")
	if err := json.Unmarshal(data, recv); err != nil {
		out.Error("Request.onIpAddr: invalid json data.")
		return
	}
	if recv.Name == "" {
		recv.Name = recv.Tenant
	}
	if recv.Name == "" {
		recv.Name = "default"
	}
	n := cache.Network.Get(recv.Name)
	if n == nil {
		out.Error("Request.onIpAddr: invalid network %s.", recv.Name)
		return
	}
	out.Cmd("Request.onIpAddr: find %s", n)
	p := cache.Point.Get(client.String())
	if p == nil {
		out.Error("Request.onIpAddr: point notFound")
		return
	}
	resp := &models.Network{
		Name:    n.Name,
		IfAddr:  recv.IfAddr,
		Netmask: recv.Netmask,
		Routes:  n.Routes,
	}
	lease := r.findLease(recv.IfAddr, p, n)
	if lease != nil {
		resp.IfAddr = lease.Address
		resp.Netmask = n.Netmask
		resp.Routes = n.Routes
	} else {
		resp.IfAddr = "169.254.0.0"
		resp.Netmask = n.Netmask
		if resp.Netmask == "" {
			resp.Netmask = "255.255.0.0"
		}
		resp.Routes = n.Routes
	}
	out.Cmd("Request.onIpAddr: resp %s", resp)
	if respStr, err := json.Marshal(resp); err == nil {
		m := libol.NewControlFrame(libol.IpAddrResp, respStr)
		_ = client.WriteMsg(m)
	}
	out.Info("Request.onIpAddr: %s", resp.IfAddr)
}

func (r *Request) onLeave(client libol.SocketClient, data []byte) {
	out := client.Out()
	out.Info("Request.onLeave")
	r.master.OffClient(client)
}
