package cswitch

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/luscis/openlan/pkg/cache"
	co "github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/models"
	nl "github.com/vishvananda/netlink"
)

type NextGroup struct {
	Network    string
	cfg        map[string]co.NextGroup
	strategies map[string]NextGroupStrategy
	out        *libol.SubLogger
}

func NewNextGroup(network string, cfg map[string]co.NextGroup) *NextGroup {
	strategies := make(map[string]NextGroupStrategy, 32)
	for key, ng := range cfg {
		strategy := newCheckStrategy(key, network, ng)

		if strategy != nil {
			strategies[key] = strategy
		}
	}

	return &NextGroup{
		Network:    network,
		cfg:        cfg,
		strategies: strategies,
		out:        libol.NewSubLogger("nextgroup"),
	}
}

func newCheckStrategy(name string, network string, cfg co.NextGroup) NextGroupStrategy {
	switch cfg.Check {
	case "ping":
		return NewPingStrategy(name, network, &cfg)
	}
	return nil
}

func (ng *NextGroup) Start() {
	ng.out.Info("NextGroup.Start: nextgroup, ng strategies size: %d", len(ng.strategies))
	if len(ng.strategies) > 0 {
		for _, checker := range ng.strategies {
			checker.Start()
		}
	}
}

func (ng *NextGroup) Stop() {
	if len(ng.strategies) > 0 {
		for _, strategy := range ng.strategies {
			strategy.Stop()
		}
	}
}

// for add nextgrou dynamicly
func (ng *NextGroup) AddNextGroup(name string, cfg co.NextGroup) {
	if _, ok := ng.strategies[name]; ok {
		ng.out.Error("NextGroup.addNextGroup: checker already exists %s", name)
		return
	}

	strategy := newCheckStrategy(name, ng.Network, cfg)
	if strategy != nil {
		ng.strategies[name] = strategy
	} else {
		ng.out.Error("NextGroup.AddNextGroup: don't support this strategy %s", name)
	}
	strategy.Start()
}

// for del nextgrou dynamicly
func (ng *NextGroup) DelNextGroup(name string, cfg co.NextGroup) {
	if strategy, ok := ng.strategies[name]; !ok {
		ng.out.Error("NextGroup.addNextGroup: checker not exists %s", name)
		return
	} else {
		if strategy.HasRoute() {
			ng.out.Error("NextGroup.delNextGroup: checker has route %s", name)
			return
		}
		strategy.Stop()
		delete(ng.strategies, name)
	}
}

func (ng *NextGroup) LoadRoute(nextgroup string, nlr *nl.Route) {
	if strategy, ok := ng.strategies[nextgroup]; ok {
		ng.out.Debug("NextGroup.loadRoute: %v", nlr)
		strategy.LoadRoute(nlr)
	} else {
		ng.out.Error("NextGroup.loadRoute: checker not found %s", nextgroup)
	}
}

func (ng *NextGroup) UnloadRoute(nextgroup string, nlr *nl.Route) {
	if strategy, ok := ng.strategies[nextgroup]; ok {
		ng.out.Debug("NextGroup.unloadRoute: %v", nlr)
		strategy.UnloadRoute(nlr)
	} else {
		ng.out.Error("NextGroup.unloadRoute: checker not found %s", nextgroup)
	}
}

type NextGroupStrategy interface {
	Name() string
	Check([]string) []co.MultiPath
	Start()
	Stop()
	ReloadRoute()
	LoadRoute(route *nl.Route)
	UnloadRoute(route *nl.Route)
	HasRoute() bool
}

type NextGroupStrategyImpl struct {
	Network string
	routes  []*nl.Route
	cfg     *co.NextGroup
	out     *libol.SubLogger
}

func (c *NextGroupStrategyImpl) Name() string {
	return "common"
}

func (c *NextGroupStrategyImpl) Start() {

}

func (c *NextGroupStrategyImpl) Stop() {

}

func (c *NextGroupStrategyImpl) HasRoute() bool {
	return len(c.routes) > 0
}

func (c *NextGroupStrategyImpl) Check(ipList []string) []co.MultiPath {
	return nil
}

func (c *NextGroupStrategyImpl) UpdateAvailableNexthops(mp []co.MultiPath) bool {
	if c.cfg.Mode == "active-backup" {
		var newPath co.MultiPath
		if len(mp) > 0 {
			newPath = mp[0]
		}
		var oldPath co.MultiPath
		if len(c.cfg.AvailableNextHop) > 0 {
			oldPath = c.cfg.AvailableNextHop[0]
		}
		if !newPath.CompareEqual(oldPath) {
			c.cfg.AvailableNextHop = []co.MultiPath{newPath}
			c.out.Debug("NextGroupStrategyImpl.UpdateAvailableNexthops: final available nexthops %v", c.cfg.AvailableNextHop)
			return true
		}

	} else if c.cfg.Mode == "load-balance" {
		if !compareMultiPaths(mp, c.cfg.AvailableNextHop) {
			c.cfg.AvailableNextHop = mp
			c.out.Debug("NextGroupStrategyImpl.UpdateAvailableNexthops: final available nexthops %v", c.cfg.AvailableNextHop)

			return true
		}
	} else {
		c.cfg.AvailableNextHop = []co.MultiPath{}
		c.out.Debug("NextGroupStrategyImpl.UpdateAvailableNexthops: final available nexthops %v", c.cfg.AvailableNextHop)
		//ignore
		return true
	}
	return false
}

func (c *NextGroupStrategyImpl) ReloadRoute() {
	c.out.Debug("NextGroupStrategyImpl.ReloadRoute: route reload %d", len(c.routes))
	for _, rt := range c.routes {
		c.updateRoute(rt)
	}
}

func (c *NextGroupStrategyImpl) modelMultiPath() []models.MultiPath {
	var modelMultiPath []models.MultiPath
	for _, mp := range c.cfg.AvailableNextHop {
		modelMultiPath = append(modelMultiPath, models.MultiPath{
			NextHop: mp.NextHop,
			Weight:  mp.Weight,
		})
	}

	return modelMultiPath
}

func (c *NextGroupStrategyImpl) buildNexthopInfos() []*nl.NexthopInfo {
	multiPath := make([]*nl.NexthopInfo, 0, len(c.cfg.AvailableNextHop))

	if len(c.cfg.AvailableNextHop) > 0 {
		for _, mr := range c.cfg.AvailableNextHop {
			nxhe := &nl.NexthopInfo{
				Hops: mr.Weight,
				Gw:   net.ParseIP(mr.NextHop),
			}
			multiPath = append(multiPath, nxhe)
		}
	}

	return multiPath
}

func (c *NextGroupStrategyImpl) updateRoute(nlr *nl.Route) {
	c.out.Debug("NextGroupStrategyImpl.updateRoute: %v ", nlr)
	multiPath := c.buildNexthopInfos()
	modelMultiPath := c.modelMultiPath()

	nlr.MultiPath = multiPath

	cache.Network.UpdateRoute(c.Network, co.PrefixRoute{
		Prefix: nlr.Dst.String(),
	}, func(obj *models.Route) {
		obj.MultiPath = modelMultiPath
	})

	promise := libol.NewPromise()
	promise.Go(func() error {
		if err := nl.RouteReplace(nlr); err != nil {
			c.out.Warn("NextGroupStrategyImpl.updateRoute: %v %s", nlr, err)
			return err
		}
		c.out.Info("NextGroupStrategyImpl.updateRoute: %v success", nlr.String())
		return nil
	})
}

func (c *NextGroupStrategyImpl) LoadRoute(nlr *nl.Route) {
	c.out.Debug("NextGroupStrategyImpl.LoadRoute: %v", nlr)
	c.routes = append(c.routes, nlr)
	nlr.MultiPath = c.buildNexthopInfos()
	nlr.Gw = nil
	c.updateRoute(nlr)
}

func (c *NextGroupStrategyImpl) UnloadRoute(rt *nl.Route) {
	c.out.Debug("NextGroupStrategyImpl.UnLoadRoute: %v", rt)
	//find route in routes
	var nlr *nl.Route
	for i, r := range c.routes {
		if r.Dst == rt.Dst && r.Table == rt.Table {
			nlr = r
			c.routes = append(c.routes[:i], c.routes[i+1:]...)
			break
		}
	}

	if nlr != nil {
		if err := nl.RouteDel(nlr); err != nil {
			c.out.Warn("NextGroupStrategyImpl.UnLoadRoute: %s", err)
			return
		}
	}

}

type PingCheckResult struct {
	Ip         string
	AvgLatency float64
	PacketLoss int
}

type PingCheckStrategy struct {
	*NextGroupStrategyImpl
	CfgName    string
	Running    bool
	PingParams *co.PingParams
	out        *libol.SubLogger
}

func NewPingStrategy(name string, network string, cfg *co.NextGroup) *PingCheckStrategy {
	return &PingCheckStrategy{
		CfgName: name,
		NextGroupStrategyImpl: &NextGroupStrategyImpl{
			Network: network,
			cfg:     cfg,
			out:     libol.NewSubLogger(cfg.Check + "_common_" + name),
		},
		PingParams: &cfg.Ping,
		out:        libol.NewSubLogger(cfg.Check + "_" + name),
	}
}

func (pc *PingCheckStrategy) Name() string {
	return "ping"
}

func filter(results []PingCheckResult, condition func(PingCheckResult) bool) []PingCheckResult {
	var filteredResults []PingCheckResult
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
func (pc *PingCheckStrategy) Check(ipList []string) []co.MultiPath {
	count := pc.PingParams.Count
	loss := pc.PingParams.Loss

	pc.out.Debug("PingCheckStrategy.Check: start check ips")
	var resultIps []PingCheckResult
	for _, ip := range ipList {
		avgLatency, packetLoss, err := pc.ping(ip, count)
		if err != nil {
			continue
		}
		resultIps = append(resultIps, PingCheckResult{
			Ip:         ip,
			AvgLatency: avgLatency,
			PacketLoss: packetLoss,
		})
	}
	// filter loss
	filterResultIps := filter(resultIps, func(result PingCheckResult) bool {
		return result.PacketLoss <= loss
	})

	sort.Slice(filterResultIps, func(i, j int) bool {
		if filterResultIps[i].PacketLoss != filterResultIps[j].PacketLoss {
			return filterResultIps[i].PacketLoss < filterResultIps[j].PacketLoss
		}
		return filterResultIps[i].AvgLatency < filterResultIps[j].AvgLatency
	})

	var sortedIPs []co.MultiPath
	for _, result := range filterResultIps {
		sortedIPs = append(sortedIPs, co.MultiPath{
			NextHop: result.Ip,
			Weight:  1,
		})
		pc.out.Debug("PingCheckStrategy.Check: available ip : %s , rtt:%.4f, loss: %d ", result.Ip, result.AvgLatency, result.PacketLoss)
	}

	return sortedIPs
}

func (pc *PingCheckStrategy) updateNextHop() {
	ipList := pc.Check(pc.cfg.NextHop)
	if pc.UpdateAvailableNexthops(ipList) {
		pc.ReloadRoute()
	}
}

func (pc *PingCheckStrategy) update() {
	frequency := pc.cfg.Ping.CheckFrequency

	if frequency <= 0 {
		frequency = 10
	}
	//wait tun device start
	time.Sleep(time.Second * time.Duration(2))
	for pc.Running = true; pc.Running; {
		pc.updateNextHop()
		time.Sleep(time.Second * time.Duration(frequency))
	}
}

func (pc *PingCheckStrategy) Start() {
	libol.Go(pc.update)
}

func (pc *PingCheckStrategy) Stop() {
	pc.Running = false
}

func (pc *PingCheckStrategy) ping(ip string, count int) (float64, int, error) {
	pingPath, err := exec.LookPath("ping")
	if err != nil {
		pc.out.Warn("PingCheckStrategy.Ping: cmd not found :", err)
	}

	output, err := libol.Exec(pingPath, ip, "-c", strconv.Itoa(count))
	if err != nil {
		pc.out.Debug("PingCheckStrategy.Ping: exec ping ip: %s, error: %s", ip, err)
		return 0, 0, err
	}

	avgLatency := pc.extractAvgLatency(output)

	packetLossRate, err := pc.extractPacketLoss(output)
	if err != nil {
		pc.out.Error("PingCheckStrategy.Ping:parse loss error : %s", err)
		return 0, 0, err
	}
	packetLoss := int(packetLossRate * float64(count) / 100)

	pc.out.Debug("PingCheckStrategy.Ping: ping ip[%s], rtt:%.4f, loss: %.f%%", ip, avgLatency, packetLossRate)
	return avgLatency, packetLoss, nil
}

func (pc *PingCheckStrategy) extractAvgLatency(outputStr string) float64 {
	pattern := `rtt min/avg/max/mdev = (\d+\.*\d*)/(\d+\.*\d*)/(\d+\.*\d*)/(\d+\.*\d*) ms`

	re := regexp.MustCompile(pattern)
	subMatches := re.FindStringSubmatch(outputStr)
	if len(subMatches) != 5 {
		pc.out.Error("PingCheckStrategy.Ping: Cannot extract average delay.")
		return 0
	}

	avgLatencyStr := subMatches[2]

	avgLatency, err := strconv.ParseFloat(avgLatencyStr, 64)
	if err != nil {
		pc.out.Error("PingCheckStrategy.Ping: parse float error : %s", err)
		return 0
	}
	return avgLatency
}

func (pc *PingCheckStrategy) extractPacketLoss(outputStr string) (float64, error) {
	re := regexp.MustCompile(`(\d+)% packet loss`)
	match := re.FindStringSubmatch(outputStr)
	if len(match) < 2 {
		return 0, fmt.Errorf("PingCheckStrategy.Ping: packet loss parse error")
	}
	lossRate, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, err
	}
	return lossRate, nil
}
