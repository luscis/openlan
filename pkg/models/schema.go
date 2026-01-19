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
		State:     o.GetState(),
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
