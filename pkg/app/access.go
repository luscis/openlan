package app

import (
	"encoding/json"
	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type Access struct {
	success int
	failed  int
	master  Master
}

func NewAccess(m Master) *Access {
	return &Access{
		master: m,
	}
}

func (p *Access) OnFrame(client libol.SocketClient, frame *libol.FrameMessage) error {
	out := client.Out()
	if out.Has(libol.LOG) {
		out.Log("Access.OnFrame %s.", frame)
	}
	if frame.IsControl() {
		action, params := frame.CmdAndParams()
		out.Debug("Access.OnFrame: %s", action)
		switch action {
		case libol.LoginReq:
			if err := p.handleLogin(client, params); err != nil {
				out.Error("Access.OnFrame: %s", err)
				m := libol.NewControlFrame(libol.LoginResp, []byte(err.Error()))
				_ = client.WriteMsg(m)
				//client.Close()
				return err
			}
			m := libol.NewControlFrame(libol.LoginResp, []byte("okay"))
			_ = client.WriteMsg(m)
		}
		//If instruct is not login and already auth, continue to process.
		if client.Have(libol.ClAuth) {
			return nil
		}
	}
	//Dropped all frames if not auth.
	if !client.Have(libol.ClAuth) {
		out.Debug("Access.OnFrame: unAuth")
		return libol.NewErr("unAuth client.")
	}
	return nil
}

func (p *Access) handleLogin(client libol.SocketClient, data []byte) error {
	out := client.Out()
	out.Debug("Access.handleLogin: %s", data)
	if client.Have(libol.ClAuth) {
		out.Warn("Access.handleLogin: already auth")
		return nil
	}
	user := &models.User{}
	if err := json.Unmarshal(data, user); err != nil {
		return libol.NewErr("Invalid json data.")
	}
	user.Update()
	out.Info("Access.handleLogin: %s on %s", user.Id(), user.Alias)
	if now, _ := cache.User.Check(user); now != nil {
		if now.Role != "admin" && now.Last != nil {
			// To offline lastly client if guest.
			p.master.OffClient(now.Last)
		}
		p.success++
		now.Last = client
		client.SetStatus(libol.ClAuth)
		out.Info("Access.handleLogin: success")
		_ = p.onAuth(client, user)
		return nil
	}
	p.failed++
	client.SetStatus(libol.ClUnAuth)
	return libol.NewErr("Auth failed.")
}

func (p *Access) onAuth(client libol.SocketClient, user *models.User) error {
	out := client.Out()
	if !client.Have(libol.ClAuth) {
		return libol.NewErr("not auth.")
	}
	out.Info("Access.onAuth")
	dev, err := p.master.NewTap(user.Network)
	if err != nil {
		return err
	}
	out.Info("Access.onAuth: on >>> %s <<<", dev.Name())
	proto := p.master.Protocol()
	m := models.NewPoint(client, dev, proto)
	m.SetUser(user)
	// free point has same uuid.
	if om := cache.Point.GetByUUID(m.UUID); om != nil {
		out.Info("Access.onAuth: OffClient %s", om.Client)
		p.master.OffClient(om.Client)
	}
	client.SetPrivate(m)
	cache.Point.Add(m)
	libol.Go(func() {
		p.master.ReadTap(dev, func(f *libol.FrameMessage) error {
			if err := client.WriteMsg(f); err != nil {
				p.master.OffClient(client)
				return err
			}
			return nil
		})
	})
	return nil
}

func (p *Access) Stats() (success, failed int) {
	return p.success, p.failed
}
