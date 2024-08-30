package cswitch

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	"github.com/luscis/openlan/pkg/schema"
	nl "github.com/vishvananda/netlink"
)

type FindHop struct {
	name    string
	cfg     *co.Network
	drivers map[string]FindHopDriver
	out     *libol.SubLogger
	lock    sync.RWMutex
}

func NewFindHop(name string, cfg *co.Network) *FindHop {
	drivers := make(map[string]FindHopDriver, 32)
	for key, ng := range cfg.FindHop {
		drv := newCheckDriver(key, name, ng)
		if drv != nil {
			drivers[key] = drv
		}
	}
	return &FindHop{
		name:    name,
		cfg:     cfg,
		drivers: drivers,
		out:     libol.NewSubLogger("findhop"),
	}
}

func newCheckDriver(name string, network string, cfg *co.FindHop) FindHopDriver {
	switch cfg.Check {
	case "ping":
		return NewPingDriver(name, network, cfg)
	}
	return nil
}

func (ng *FindHop) Start() {
	ng.out.Info("FindHop.Start: drivers size: %d", len(ng.drivers))
	if len(ng.drivers) > 0 {
		for _, checker := range ng.drivers {
			checker.Start()
		}
	}
}

func (ng *FindHop) Stop() {
	if len(ng.drivers) > 0 {
		for _, driver := range ng.drivers {
			driver.Stop()
		}
	}
}

// for add findhop dynamicly
func (ng *FindHop) addHop(name string, cfg *co.FindHop) error {
	if _, ok := ng.drivers[name]; ok {
		ng.out.Error("FindHop.addHop: checker already exists %s", name)
		return nil
	}
	driver := newCheckDriver(name, ng.name, cfg)
	if driver == nil {
		return libol.NewErr("FindHop.AddHop: don't support this driver %s", name)
	}
	ng.drivers[name] = driver
	driver.Start()
	return nil
}

// for del findhop dynamicly
func (ng *FindHop) removeHop(name string) error {
	if driver, ok := ng.drivers[name]; !ok {
		ng.out.Error("FindHop.addHop: checker not exists %s", name)
		return nil
	} else {
		if driver.HasRoute() {
			return libol.NewErr("FindHop.delHop: checker has route %s", name)
		}
		driver.Stop()
		delete(ng.drivers, name)
	}
	return nil
}

func (ng *FindHop) LoadHop(findhop string, nlr *nl.Route) {
	ng.lock.RLock()
	defer ng.lock.RUnlock()
	if driver, ok := ng.drivers[findhop]; ok {
		ng.out.Info("FindHop.LoadHop: %s via %s", nlr.String(), findhop)
		driver.LoadRoute(nlr)
	} else {
		ng.out.Error("FindHop.LoadHop: checker not found %s", findhop)
	}
}

func (ng *FindHop) UnloadHop(findhop string, nlr *nl.Route) {
	ng.lock.RLock()
	defer ng.lock.RUnlock()
	if driver, ok := ng.drivers[findhop]; ok {
		ng.out.Info("FindHop.UnloadHop: %s via %s", nlr.String(), findhop)
		driver.UnloadRoute(nlr)
	} else {
		ng.out.Error("FindHop.UnloadHop: checker not found %s", findhop)
	}
}

func (ng *FindHop) AddHop(data schema.FindHop) error {
	ng.lock.Lock()
	defer ng.lock.Unlock()
	cc := &co.FindHop{
		Name:    data.Name,
		Mode:    data.Mode,
		NextHop: strings.Split(data.NextHop, ","),
		Check:   data.Check,
	}
	cc.Correct()
	if ng.cfg.AddFindHop(cc) {
		return ng.addHop(data.Name, cc)
	}
	return nil
}

func (ng *FindHop) DelHop(data schema.FindHop) error {
	ng.lock.Lock()
	defer ng.lock.Unlock()
	cc := &co.FindHop{
		Name: data.Name,
	}
	if err := ng.removeHop(data.Name); err == nil {
		ng.cfg.DelFindHop(cc)
		return nil
	} else {
		return err
	}
}

func (ng *FindHop) ListHop(call func(obj schema.FindHop)) {
	ng.lock.RLock()
	defer ng.lock.RUnlock()
	for name, drv := range ng.drivers {
		cc := drv.Config()
		avas := make([]string, 0)
		for _, ava := range cc.Available {
			avas = append(avas, ava.NextHop)
		}
		call(schema.FindHop{
			Name:      name,
			Mode:      cc.Mode,
			NextHop:   strings.Join(cc.NextHop, ","),
			Check:     cc.Check,
			Available: strings.Join(avas, ","),
		})
	}
}

func (ng *FindHop) SaveHop() {
	ng.lock.RLock()
	defer ng.lock.RUnlock()
	ng.cfg.SaveFindHop()
}

type FindHopDriver interface {
	Name() string
	Check([]string) []co.MultiPath
	Start()
	Stop()
	ReloadRoute()
	LoadRoute(route *nl.Route)
	UnloadRoute(route *nl.Route)
	HasRoute() bool
	Config() *co.FindHop
}

type FindHopImpl struct {
	Network string
	routes  []*nl.Route
	cfg     *co.FindHop
	out     *libol.SubLogger
}

func (c *FindHopImpl) Name() string {
	return "common"
}

func (c *FindHopImpl) Start() {
}

func (c *FindHopImpl) Stop() {
}

func (c *FindHopImpl) HasRoute() bool {
	return len(c.routes) > 0
}

func (c *FindHopImpl) Check(ipList []string) []co.MultiPath {
	return nil
}

func (c *FindHopImpl) UpdateAvailable(mp []co.MultiPath) bool {
	if c.cfg.Mode == "load-balance" {
		if !compareMultiPaths(mp, c.cfg.Available) {
			c.cfg.Available = mp
			c.out.Info("FindHopImpl.UpdateAvailable: available %v", c.cfg.Available)
			return true
		}
	} else {
		var newPath co.MultiPath
		if len(mp) > 0 {
			newPath = mp[0]
		}
		var oldPath co.MultiPath
		if len(c.cfg.Available) > 0 {
			oldPath = c.cfg.Available[0]
		}
		if !newPath.CompareEqual(oldPath) {
			c.cfg.Available = []co.MultiPath{newPath}
			c.out.Info("FindHopImpl.UpdateAvailable: available %v", c.cfg.Available)
			return true
		}
	}
	return false
}

func (c *FindHopImpl) ReloadRoute() {
	c.out.Debug("FindHopImpl.ReloadRoute: route reload %d", len(c.routes))
	for _, rt := range c.routes {
		c.updateRoute(rt)
	}
}

func (c *FindHopImpl) modelMultiPath() []models.MultiPath {
	var modelMultiPath []models.MultiPath
	for _, mp := range c.cfg.Available {
		modelMultiPath = append(modelMultiPath, models.MultiPath{
			NextHop: mp.NextHop,
			Weight:  mp.Weight,
		})
	}
	return modelMultiPath
}

func (c *FindHopImpl) buildNexthopInfos() []*nl.NexthopInfo {
	multiPath := make([]*nl.NexthopInfo, 0, len(c.cfg.Available))
	if len(c.cfg.Available) > 0 {
		for _, mr := range c.cfg.Available {
			nxhe := &nl.NexthopInfo{
				Hops: mr.Weight,
				Gw:   net.ParseIP(mr.NextHop),
			}
			multiPath = append(multiPath, nxhe)
		}
	}
	return multiPath
}

func (c *FindHopImpl) updateRoute(nlr *nl.Route) {
	c.out.Debug("FindHopImpl.updateRoute: %v ", nlr)
	multiPath := c.buildNexthopInfos()

	nlr.MultiPath = multiPath

	promise := libol.NewPromise()
	promise.Go(func() error {
		if err := nl.RouteReplace(nlr); err != nil {
			c.out.Warn("FindHopImpl.updateRoute: %s %s", nlr.String(), err)
			return err
		}
		c.out.Info("FindHopImpl.updateRoute: %s success", nlr.String())
		return nil
	})
}

func (c *FindHopImpl) LoadRoute(nlr *nl.Route) {
	c.out.Debug("FindHopImpl.LoadRoute: %v", nlr)
	c.routes = append(c.routes, nlr)
	nlr.MultiPath = c.buildNexthopInfos()
	nlr.Gw = nil
	if len(nlr.MultiPath) == 0 {
		c.out.Debug("ignored if no nexthop")
	} else {
		c.updateRoute(nlr)
	}
}

func (c *FindHopImpl) UnloadRoute(rt *nl.Route) {
	c.out.Debug("FindHopImpl.UnLoadRoute: %v", rt)
	//find route in routes
	var nlr *nl.Route
	for i, r := range c.routes {
		if r.Dst.String() == rt.Dst.String() && r.Table == rt.Table {
			nlr = r
			c.routes = append(c.routes[:i], c.routes[i+1:]...)
			break
		}
	}

	if nlr != nil {
		if err := nl.RouteDel(nlr); err != nil {
			c.out.Warn("FindHopImpl.UnLoadRoute: %s", err)
			return
		}
	}
}

func (c *FindHopImpl) Config() *co.FindHop {
	return c.cfg
}

type PingResult struct {
	Ip      string
	Latency float64
	Loss    int
}

type PingDriver struct {
	*FindHopImpl
	CfgName    string
	Running    bool
	PingParams *co.PingParams
}

func NewPingDriver(name string, network string, cfg *co.FindHop) *PingDriver {
	return &PingDriver{
		CfgName: name,
		FindHopImpl: &FindHopImpl{
			Network: network,
			cfg:     cfg,
			out:     libol.NewSubLogger(cfg.Check + "_" + name),
		},
		PingParams: &cfg.Params,
	}
}

func (pc *PingDriver) Name() string {
	return "ping"
}

func filter(results []PingResult, condition func(PingResult) bool) []PingResult {
	var filteredResults []PingResult
	for _, result := range results {
		if condition(result) {
			filteredResults = append(filteredResults, result)
		}
	}
	return filteredResults
}

func compareMultiPaths(a []co.MultiPath, b []co.MultiPath) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !a[i].CompareEqual(b[i]) {
			return false
		}
	}
	return true
}

// check the ipList and return the available NextHops
func (pc *PingDriver) Check(ipList []string) []co.MultiPath {
	count := pc.PingParams.Count
	loss := pc.PingParams.Loss

	pc.out.Debug("PingDriver.Check: start check ips")
	var resultIps []PingResult
	for _, ip := range ipList {
		avgLatency, packetLoss, err := pc.ping(ip, count)
		if err != nil {
			continue
		}
		resultIps = append(resultIps, PingResult{
			Ip:      ip,
			Latency: avgLatency,
			Loss:    packetLoss,
		})
	}
	// filter loss
	filterResultIps := filter(resultIps, func(result PingResult) bool {
		return result.Loss <= loss
	})

	sort.Slice(filterResultIps, func(i, j int) bool {
		ii := filterResultIps[i]
		jj := filterResultIps[j]
		if ii.Loss != jj.Loss {
			return ii.Loss < jj.Loss
		}
		return ii.Ip < jj.Ip
	})

	var sortedIPs []co.MultiPath
	for _, result := range filterResultIps {
		sortedIPs = append(sortedIPs, co.MultiPath{
			NextHop: result.Ip,
			Weight:  1,
		})
		pc.out.Debug("PingDriver.Check: available %s loss: %d ", result.Ip, result.Loss)
	}

	return sortedIPs
}

func (pc *PingDriver) updateNextHop() {
	ipList := pc.Check(pc.cfg.NextHop)
	if pc.UpdateAvailable(ipList) {
		pc.ReloadRoute()
	}
}

func (pc *PingDriver) update() {
	frequency := pc.cfg.Params.Interval
	if frequency <= 0 {
		frequency = 5
	}

	//wait tun device start
	time.Sleep(time.Second * time.Duration(2))
	for pc.Running = true; pc.Running; {
		pc.updateNextHop()
		time.Sleep(time.Second * time.Duration(frequency))
	}
}

func (pc *PingDriver) Start() {
	libol.Go(pc.update)
}

func (pc *PingDriver) Stop() {
	pc.Running = false
}

func (pc *PingDriver) ping(ip string, count int) (float64, int, error) {
	pingPath, err := exec.LookPath("ping")
	if err != nil {
		pc.out.Warn("PingDriver.Ping: cmd not found :", err)
	}

	output, err := libol.Exec(pingPath, ip, "-c", strconv.Itoa(count))
	if err != nil {
		pc.out.Debug("PingDriver.Ping: exec ping %s, error: %s", ip, err)
		return 0, 0, err
	}

	avgLatency := pc.extractLatency(output)
	LossRate, err := pc.extractLoss(output)
	if err != nil {
		pc.out.Error("PingDriver.Ping:parse loss error : %s", err)
		return 0, 0, err
	}

	packetLoss := int(LossRate * float64(count) / 100)
	pc.out.Debug("PingDriver.Ping: ping ip[%s] loss:%.f%", ip, avgLatency, LossRate)
	return avgLatency, packetLoss, nil
}

func (pc *PingDriver) extractLatency(outputStr string) float64 {
	pattern := `rtt min/avg/max/mdev = (\d+\.*\d*)/(\d+\.*\d*)/(\d+\.*\d*)/(\d+\.*\d*) ms`

	re := regexp.MustCompile(pattern)
	subMatches := re.FindStringSubmatch(outputStr)
	if len(subMatches) != 5 {
		pc.out.Error("PingDriver.Ping: Cannot extract average delay.")
		return 0
	}

	avgLatencyStr := subMatches[2]
	avgLatency, err := strconv.ParseFloat(avgLatencyStr, 64)
	if err != nil {
		pc.out.Error("PingDriver.Ping: parse float error : %s", err)
		return 0
	}
	return avgLatency
}

func (pc *PingDriver) extractLoss(outputStr string) (float64, error) {
	re := regexp.MustCompile(`(\d+)% packet loss`)
	match := re.FindStringSubmatch(outputStr)
	if len(match) < 2 {
		return 0, fmt.Errorf("PingDriver.Ping: packet loss parse error")
	}

	lossRate, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}
	return lossRate, nil
}
