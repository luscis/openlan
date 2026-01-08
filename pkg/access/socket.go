package access

import (
	"encoding/json"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
)

type SocketWorkerListener struct {
	OnClose   func(w *SocketWorker) error
	OnSuccess func(w *SocketWorker) error
	OnIpAddr  func(w *SocketWorker, n *models.Network) error
	ReadAt    func(frame *libol.FrameMessage) error
}

const (
	rtLast      = "lastAt"   // record time last frame received or connected.
	rtConnected = "connAt"   // record last connected time.
	rtReConnect = "reconnAt" // record time when triggered reconnected.
	rtSuccess   = "succAt"   // record success time when login.
	rtSleeps    = "sleeps"   // record times to control connecting delay.
	rtClosed    = "clsAt"    // close time
	rtLive      = "liveAt"   // record received pong frame time.
	rtIpAddr    = "addrAt"   // record last receive ipAddr message after success.
	rtConnects  = "conns"    // record times of reconnecting
	rtLatency   = "latency"  // latency by ping.
)

type SocketWorker struct {
	// private
	listener   SocketWorkerListener
	client     libol.SocketClient
	lock       sync.Mutex
	user       *models.User
	network    *models.Network
	routes     map[string]*models.Route
	keepalive  KeepAlive
	done       chan bool
	ticker     *time.Ticker
	pinCfg     *config.Access
	eventQueue chan *WorkerEvent
	writeQueue chan *libol.FrameMessage
	jobber     []jobTimer
	record     *libol.SafeStrInt64
	out        *libol.SubLogger
	wlFrame    *libol.FrameMessage // Last frame from write.
}

func NewSocketWorker(client libol.SocketClient, c *config.Access) *SocketWorker {
	module := client.String() + "|" + c.Network
	t := &SocketWorker{
		client:     client,
		network:    models.NewNetwork(c.Network, c.Interface.Address),
		routes:     make(map[string]*models.Route, 64),
		record:     libol.NewSafeStrInt64(),
		done:       make(chan bool, 2),
		ticker:     time.NewTicker(2 * time.Second),
		pinCfg:     c,
		eventQueue: make(chan *WorkerEvent, 32),
		writeQueue: make(chan *libol.FrameMessage, c.Queue.SockWr),
		jobber:     make([]jobTimer, 0, 32),
		out:        libol.NewSubLogger(module),
	}
	t.user = &models.User{
		Alias:    c.Alias,
		Name:     c.Username,
		Password: c.Password,
		Network:  c.Network,
		System:   runtime.GOOS,
	}
	t.keepalive = KeepAlive{
		Interval: 15,
		LastTime: time.Now().Unix(),
	}
	return t
}

func (t *SocketWorker) sleepNow() int64 {
	sleeps := t.record.Get(rtSleeps)
	return sleeps * 5
}

func (t *SocketWorker) sleepIdle() int64 {
	sleeps := t.record.Get(rtSleeps)
	if sleeps < 20 {
		t.record.Add(rtSleeps, 1)
	}
	return t.sleepNow()
}

func (t *SocketWorker) Initialize() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Initialize")
	t.client.SetMaxSize(t.pinCfg.Interface.IPMtu)
	t.client.SetListener(libol.ClientListener{
		OnConnected: func(client libol.SocketClient) error {
			t.record.Set(rtConnected, time.Now().Unix())
			t.eventQueue <- NewEvent(EvSocConed, "from socket")
			return nil
		},
		OnClose: func(client libol.SocketClient) error {
			t.record.Set(rtClosed, time.Now().Unix())
			t.eventQueue <- NewEvent(EvSocClosed, "from socket")
			return nil
		},
	})
	t.record.Set(rtLast, time.Now().Unix())
	t.record.Set(rtReConnect, time.Now().Unix())
}

func (t *SocketWorker) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Start")
	libol.Go(t.Loop)
}

func (t *SocketWorker) sendLeave(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	data := struct {
		DateTime   int64  `json:"datetime"`
		UUID       string `json:"uuid"`
		Alias      string `json:"alias"`
		Connection string `json:"connection"`
		Address    string `json:"address"`
	}{
		DateTime:   time.Now().Unix(),
		UUID:       t.user.UUID,
		Alias:      t.user.Alias,
		Address:    t.client.LocalAddr(),
		Connection: t.client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.leave: left: %s", body)
	m := libol.NewControlFrame(libol.LeftReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

func (t *SocketWorker) leave() {
	t.out.Info("SocketWorker.leave")
	if err := t.sendLeave(t.client); err != nil {
		t.out.Error("SocketWorker.leave: %s", err)
	}
}

func (t *SocketWorker) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.out.Info("SocketWorker.Stop")
	t.leave()
	t.client.Terminal()
	t.done <- true
	t.client = nil
	t.ticker.Stop()
}

func (t *SocketWorker) close() {
	if t.client != nil {
		t.client.Close()
	}
}

func (t *SocketWorker) connect() error {
	t.out.Warn("SocketWorker.connect: %s", t.client)
	t.client.Close()
	s := t.client.Status()
	if s != libol.ClInit {
		t.out.Warn("SocketWorker.connect: %s %s", t.client, s)
		t.client.SetStatus(libol.ClInit)
	}
	t.record.Add(rtConnects, 1)
	if err := t.client.Connect(); err != nil {
		t.out.Error("SocketWorker.connect: %s %s", t.client, err)
		return err
	}
	return nil
}

func (t *SocketWorker) reconnect() {
	if t.isStopped() {
		return
	}
	t.record.Set(rtReConnect, time.Now().Unix())
	job := jobTimer{
		Time: time.Now().Unix() + t.sleepIdle(),
		Call: func() error {
			t.out.Debug("SocketWorker.reconnect: on jobber")
			rtConn := t.record.Get(rtConnected)
			rtReCon := t.record.Get(rtReConnect)
			rtLast := t.record.Get(rtLast)
			rtLive := t.record.Get(rtLive)
			if rtConn >= rtReCon { // already connected after.
				t.out.Cmd("SocketWorker.reconnect: dissed by connected")
				return nil
			}
			t.out.Info("SocketWorker.reconnect: l: %d a: %d", rtLast, rtLive)
			t.out.Info("SocketWorker.reconnect: c: %d r: %d", rtConn, rtReCon)
			return t.connect()
		},
	}
	t.jobber = append(t.jobber, job)
}

func (t *SocketWorker) sendLogin(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	body, err := json.Marshal(t.user)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.toLogin: %s", body)
	m := libol.NewControlFrame(libol.LoginReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

// toLogin request
func (t *SocketWorker) toLogin(client libol.SocketClient) error {
	if err := t.sendLogin(client); err != nil {
		t.out.Error("SocketWorker.toLogin: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) sendIpAddr(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	body, err := json.Marshal(t.network)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.toNetwork: %s", body)
	m := libol.NewControlFrame(libol.IpAddrReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

func (t *SocketWorker) canReqAddr() bool {
	if t.pinCfg.RequestAddr {
		return true
	}
	// For link, need advise ipAddr with configured address.
	if t.network.Address != "" {
		return true
	}
	return false
}

// network request
func (t *SocketWorker) toNetwork(client libol.SocketClient) error {
	if !t.canReqAddr() {
		t.out.Info("SocketWorker.toNetwork: notNeed")
		return nil
	}
	if err := t.sendIpAddr(client); err != nil {
		t.out.Error("SocketWorker.toNetwork: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) onLogin(resp []byte) error {
	if t.client.Have(libol.ClAuth) {
		t.out.Cmd("SocketWorker.onLogin: %s", resp)
		return nil
	}
	if strings.HasPrefix(string(resp), "okay") {
		t.client.SetStatus(libol.ClAuth)
		if t.listener.OnSuccess != nil {
			_ = t.listener.OnSuccess(t)
		}
		t.record.Set(rtSleeps, 0)
		t.record.Set(rtIpAddr, 0)
		t.record.Set(rtSuccess, time.Now().Unix())
		t.eventQueue <- NewEvent(EvSocSuccess, "from login")
		t.out.Info("SocketWorker.onLogin: success")
	} else {
		t.client.SetStatus(libol.ClUnAuth)
		t.out.Error("SocketWorker.onLogin: %s", resp)
	}
	return nil
}

func (t *SocketWorker) onIpAddr(resp []byte) error {
	if !t.pinCfg.RequestAddr {
		t.out.Info("SocketWorker.onIpAddr: notAllowed")
		return nil
	}
	n := &models.Network{}
	if err := json.Unmarshal(resp, n); err != nil {
		return libol.NewErr("SocketWorker.onIpAddr: invalid json data.")
	}
	t.network = n
	if t.listener.OnIpAddr != nil {
		_ = t.listener.OnIpAddr(t, n)
	}
	return nil
}

func (t *SocketWorker) onLeft(resp []byte) error {
	t.out.Info("SocketWorker.onLeft")
	t.out.Cmd("SocketWorker.onLeft: %s", resp)
	t.close()
	return nil
}

func (t *SocketWorker) onSignIn(resp []byte) error {
	t.out.Info("SocketWorker.onSignIn")
	t.out.Cmd("SocketWorker.onSignIn: %s", resp)
	t.eventQueue <- NewEvent(EvSocSignIn, "request from server")
	return nil
}

func (t *SocketWorker) onPong(resp []byte) error {
	m := &PingMsg{}
	if err := json.Unmarshal(resp, m); err != nil {
		return libol.NewErr("SocketWorker.onPong: invalid json data.")
	}
	latency := time.Now().UnixNano() - m.DateTime // ns
	t.record.Set(rtLatency, latency/1e6)          // ms
	return nil
}

// handle instruct from virtual switch
func (t *SocketWorker) onInstruct(frame *libol.FrameMessage) error {
	if !frame.IsControl() {
		return nil
	}
	action, resp := frame.CmdAndParams()
	if libol.HasLog(libol.CMD) {
		t.out.Cmd("SocketWorker.onInstruct %s %s", action, resp)
	}
	switch action {
	case libol.LoginResp:
		return t.onLogin(resp)
	case libol.IpAddrResp:
		t.record.Set(rtIpAddr, time.Now().Unix())
		return t.onIpAddr(resp)
	case libol.PongResp:
		t.record.Set(rtLive, time.Now().Unix())
		return t.onPong(resp)
	case libol.SignReq:
		return t.onSignIn(resp)
	case libol.LeftReq:
		return t.onLeft(resp)
	default:
		t.out.Warn("SocketWorker.onInstruct: %s %s", action, resp)
	}
	return nil
}

type PingMsg struct {
	DateTime   int64  `json:"datetime"`
	UUID       string `json:"uuid"`
	Alias      string `json:"alias"`
	Connection string `json:"connection"`
	Address    string `json:"address"`
}

func (t *SocketWorker) sendPing(client libol.SocketClient) error {
	if client == nil {
		return libol.NewErr("client is nil")
	}
	data := &PingMsg{
		DateTime:   time.Now().UnixNano(),
		UUID:       t.user.UUID,
		Alias:      t.user.Alias,
		Address:    t.client.LocalAddr(),
		Connection: t.client.RemoteAddr(),
	}
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	t.out.Cmd("SocketWorker.sendPing: ping= %s", body)
	m := libol.NewControlFrame(libol.PingReq, body)
	if err := client.WriteMsg(m); err != nil {
		return err
	}
	return nil
}

func (t *SocketWorker) keepAlive() {
	if !t.keepalive.Should() {
		return
	}
	t.keepalive.Update()
	if t.client.Have(libol.ClAuth) {
		// Whether ipAddr request was already? and try ipAddr?
		rtIp := t.record.Get(rtIpAddr)
		rtSuc := t.record.Get(rtSuccess)
		if t.canReqAddr() && rtIp < rtSuc {
			_ = t.toNetwork(t.client)
		}
		if err := t.sendPing(t.client); err != nil {
			t.out.Error("SocketWorker.keepAlive: %s", err)
		}
	} else {
		if err := t.sendLogin(t.client); err != nil {
			t.out.Error("SocketWorker.keepAlive: %s", err)
		}
	}
}

func (t *SocketWorker) checkJobber() {
	// travel jobber and execute it expired.
	now := time.Now().Unix()
	newTimer := make([]jobTimer, 0, 32)
	for _, t := range t.jobber {
		if now >= t.Time {
			_ = t.Call()
		} else {
			newTimer = append(newTimer, t)
		}
	}
	t.jobber = newTimer
	t.out.Debug("SocketWorker.checkJobber: %d", len(t.jobber))
}

func (t *SocketWorker) checkAlive() {
	out := int64(t.pinCfg.Timeout)
	now := time.Now().Unix()
	if now-t.record.Get(rtLast) < out || now-t.record.Get(rtLive) < out {
		return
	}
	if now-t.record.Get(rtReConnect) < out { // timeout and avoid send reconn frequently.
		t.out.Cmd("SocketWorker.checkAlive: reconn frequently")
		return
	}
	t.eventQueue <- NewEvent(EvSocRecon, "from alive check")
}

func (t *SocketWorker) doTicker() error {
	t.checkAlive()  // period to check whether alive.
	t.keepAlive()   // send ping and wait pong to keep alive.
	t.checkJobber() // period to check job whether timeout.
	return nil
}

func (t *SocketWorker) dispatch(ev *WorkerEvent) {
	t.out.Event("SocketWorker.dispatch: %v", ev)
	switch ev.Type {
	case EvSocConed:
		if t.client != nil {
			_ = t.toLogin(t.client)
			libol.Go(func() {
				t.Read(t.client)
			})
		}
	case EvSocSuccess:
		_ = t.toNetwork(t.client)
		_ = t.sendPing(t.client)
	case EvSocRecon:
		t.out.Info("SocketWorker.dispatch: %v", ev)
		t.reconnect()
	case EvSocSignIn, EvSocLogin:
		_ = t.toLogin(t.client)
	}
}

func (t *SocketWorker) Loop() {
	err := t.connect()
	if err != nil {
		t.out.Warn("SocketWorker.Loop: %s", err)
	}
	for {
		select {
		case e := <-t.eventQueue:
			t.lock.Lock()
			t.dispatch(e)
			t.lock.Unlock()
		case d := <-t.writeQueue:
			_ = t.DoWrite(d)
		case <-t.done:
			return
		case c := <-t.ticker.C:
			t.out.Log("SocketWorker.Ticker: at %s", c)
			t.lock.Lock()
			_ = t.doTicker()
			t.lock.Unlock()
		}
	}
}

func (t *SocketWorker) isStopped() bool {
	return t.client == nil || t.client.Have(libol.ClTerminal)
}

func (t *SocketWorker) Read(client libol.SocketClient) {
	for {
		data, err := client.ReadMsg()
		if err != nil {
			t.out.Error("SocketWorker.Read: %s", err)
			client.Close()
			break
		}
		if t.out.Has(libol.DEBUG) {
			t.out.Debug("SocketWorker.Read: %x", data)
		}
		if data.Size() <= 0 {
			continue
		}
		data.Decode()
		if data.IsControl() {
			t.lock.Lock()
			_ = t.onInstruct(data)
			t.lock.Unlock()
			continue
		}
		t.record.Set(rtLast, time.Now().Unix())
		if t.listener.ReadAt != nil {
			_ = t.listener.ReadAt(data)
		}
	}
	if !t.isStopped() {
		t.eventQueue <- NewEvent(EvSocRecon, "from read")
	}
}

func (t *SocketWorker) DoWrite(frame *libol.FrameMessage) error {
	if t.out.Has(libol.DEBUG) {
		t.out.Debug("SocketWorker.DoWrite: %x", frame)
	}
	t.checkAlive() // alive check immediately
	t.lock.Lock()
	if t.client == nil {
		t.lock.Unlock()
		return libol.NewErr("client is nil")
	}
	if !t.client.Have(libol.ClAuth) {
		t.out.Debug("SocketWorker.DoWrite: dropping by unAuth")
		t.lock.Unlock()
		return nil
	}
	t.lock.Unlock()
	if err := t.client.WriteMsg(frame); err != nil {
		t.out.Debug("SocketWorker.DoWrite: %s", err)
		return err
	}
	return nil
}

func (t *SocketWorker) Write(frame *libol.FrameMessage) error {
	t.writeQueue <- frame
	return nil
}

func (t *SocketWorker) Auth() (string, string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.user.Name, t.user.Password
}

func (t *SocketWorker) SetAuth(auth string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	values := strings.Split(auth, ":")
	t.user.Name = values[0]
	if len(values) > 1 {
		t.user.Password = values[1]
	}
}

func (t *SocketWorker) SetUUID(v string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.user.UUID = v
}
