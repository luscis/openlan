package app

import (
	"encoding/json"

	"github.com/luscis/openlan/pkg/cache"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/libsock"
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

func (p *Access) OnFrame(client libsock.SocketClient, frame *libsock.FrameMessage) error {
	out := client.Out()
	if out.Has(libol.LOG) {
		out.Log("Access.OnFrame %s.", frame)
	}
	if frame.IsControl() {
		action, params := frame.CmdAndParams()
		out.Debug("Access.OnFrame: %s", action)
		switch action {
		case libsock.LoginReq:
			if err := p.handleLogin(client, params); err != nil {
				out.Error("Access.OnFrame: %s", err)
				m := libsock.NewControlFrame(libsock.LoginResp, []byte(err.Error()))
				_ = client.WriteMsg(m)
				//client.Close()
				return err
			}
			m := libsock.NewControlFrame(libsock.LoginResp, []byte("okay"))
			_ = client.WriteMsg(m)
		}
		//If instruct is not login and already auth, continue to process.
		if client.Have(libsock.ClAuth) {
			return nil
		}
	}
	//Dropped all frames if not auth.
	if !client.Have(libsock.ClAuth) {
		out.Debug("Access.OnFrame: unAuth")
		return libol.NewErr("unAuth client.")
	}
	return nil
}

func (p *Access) handleLogin(client libsock.SocketClient, data []byte) error {
	out := client.Out()
	out.Debug("Access.handleLogin: %s", data)
	if client.Have(libsock.ClAuth) {
		out.Warn("Access.handleLogin: already auth")
		return nil
	}
	user := &models.User{}
	if err := json.Unmarshal(data, user); err != nil {
		return libol.NewErr("Invalid json data.")
	}
	user.Update()
	out.Info("Access.handleLogin: %s on %s", user.Id(), user.Alias)
	if now, err := cache.User.Check(user); now != nil {
		if now.Role != "admin" && now.Last != nil {
			// To offline lastly client if guest.
			p.master.OffClient(now.Last)
		}
		p.success++
		now.Last = client
		client.SetStatus(libsock.ClAuth)
		out.Info("Access.handleLogin: success")
		_ = p.onAuth(client, user)
		return nil
	} else {
		p.failed++
		client.SetStatus(libsock.ClUnAuth)
		return err
	}
}

func (p *Access) onAuth(client libsock.SocketClient, user *models.User) error {
	out := client.Out()
	if !client.Have(libsock.ClAuth) {
		return libol.NewErr("not auth.")
	}
	out.Info("Access.onAuth")
	dev, err := p.master.NewTap(user.Network)
	if err != nil {
		return err
	}
	out.Info("Access.onAuth: on >>> %s <<<", dev.Name())
	proto := client.Protocol()
	if proto == "" {
		proto = p.master.Protocol()
	}
	m := models.NewAccess(client, dev, proto)
	m.SetUser(user)
	// free point has same uuid.
	if om := cache.Access.GetByUUID(m.UUID); om != nil {
		out.Info("Access.onAuth: OffClient %s", om.Client)
		p.master.OffClient(om.Client)
	}
	client.SetPrivate(m)
	cache.Access.Add(m)
	libol.Go(func() {
		p.master.ReadTap(dev, func(f *libsock.FrameMessage) error {
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
