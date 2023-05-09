package models

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

func NewPointSchema(p *Point) schema.Point {
	client, dev := p.Client, p.Device
	sts := client.Statistics()
	return schema.Point{
		Uptime:    p.Uptime,
		UUID:      p.UUID,
		Alias:     p.Alias,
		User:      p.User,
		Protocol:  p.Protocol,
		Remote:    client.String(),
		Device:    dev.Name(),
		RxBytes:   sts[libol.CsRecvOkay],
		TxBytes:   sts[libol.CsSendOkay],
		ErrPkt:    sts[libol.CsSendError],
		State:     client.Status().String(),
		Network:   p.Network,
		AliveTime: client.AliveTime(),
		System:    p.System,
	}
}

func NewLinkSchema(l *Link) schema.Link {
	sts := l.Status()
	return schema.Link{
		UUID:      sts.UUID,
		User:      sts.User,
		Uptime:    sts.Uptime,
		Device:    sts.Device,
		Protocol:  sts.Protocol,
		Server:    sts.Remote,
		State:     sts.State,
		RxBytes:   sts.RxBytes,
		TxBytes:   sts.TxBytes,
		ErrPkt:    sts.ErrPkt,
		Network:   sts.Network,
		AliveTime: sts.AliveTime,
	}
}

func NewNeighborSchema(n *Neighbor) schema.Neighbor {
	return schema.Neighbor{
		Uptime:  n.UpTime(),
		HwAddr:  n.HwAddr.String(),
		IpAddr:  n.IpAddr.String(),
		Client:  n.Client,
		Network: n.Network,
		Device:  n.Device,
	}
}

func NewOnLineSchema(l *Line) schema.OnLine {
	return schema.OnLine{
		HitTime:    l.LastTime(),
		UpTime:     l.UpTime(),
		EthType:    l.EthType,
		IpSource:   l.IpSource.String(),
		IpDest:     l.IpDest.String(),
		IpProto:    libol.IpProto2Str(l.IpProtocol),
		PortSource: l.PortSource,
		PortDest:   l.PortDest,
	}
}

func NewUserSchema(u *User) schema.User {
	return schema.User{
		Name:     u.Name,
		Password: u.Password,
		Alias:    u.Alias,
		Network:  u.Network,
		Role:     u.Role,
		Lease:    u.Lease.Format(libol.LeaseTime),
	}
}

func SchemaToUserModel(user *schema.User) *User {
	lease, _ := libol.GetLocalTime(libol.LeaseTime, user.Lease)
	obj := &User{
		Alias:    user.Alias,
		Password: user.Password,
		Name:     user.Name,
		Network:  user.Network,
		Role:     user.Role,
		Lease:    lease,
	}
	obj.Update()
	return obj
}

func NewNetworkSchema(n *Network) schema.Network {
	sn := schema.Network{
		Name: n.Name,
		Subnet: schema.Subnet{
			IfAddr:  n.IfAddr,
			IpStart: n.IpStart,
			IpEnd:   n.IpEnd,
			Netmask: n.Netmask,
		},
		Routes: make([]schema.PrefixRoute, 0, 32),
	}
	for _, route := range n.Routes {
		sn.Routes = append(sn.Routes,
			schema.PrefixRoute{
				NextHop: route.NextHop,
				Prefix:  route.Prefix,
				Metric:  route.Metric,
				Mode:    route.Mode,
				Origin:  route.Origin,
			})
	}
	return sn
}
