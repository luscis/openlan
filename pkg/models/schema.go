package models

import (
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
)

func NewAccessSchema(p *Access) schema.Access {
	client, dev := p.Client, p.Device
	sts := client.Statistics()
	return schema.Access{
		Uptime:    p.Uptime,
		UUID:      p.UUID,
		Alias:     p.Alias,
		User:      p.User,
		Protocol:  p.Protocol,
		Remote:    client.String(),
		Device:    dev.Name(),
		RxBytes:   uint64(sts[libol.CsRecvOkay]),
		TxBytes:   uint64(sts[libol.CsSendOkay]),
		ErrPkt:    uint64(sts[libol.CsSendError]),
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
	lease, _ := libol.GetLeaseTime(user.Lease)
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
		Name:   n.Name,
		Config: n.Config,
	}
	return sn
}

func NewOutputSchema(o *Output) schema.Output {
	return schema.Output{
		Network:   o.Network,
		Protocol:  o.Protocol,
		Remote:    o.Remote,
		Fallback:  o.Fallback,
		Segment:   o.Segment,
		Device:    o.Device,
		RxBytes:   o.RxBytes,
		TxBytes:   o.TxBytes,
		Secret:    o.Secret,
		AliveTime: o.UpTime(),
	}
}
