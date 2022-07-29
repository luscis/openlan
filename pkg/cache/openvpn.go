package cache

import (
	"bufio"
	"github.com/luscis/openlan/pkg/config"
	"github.com/luscis/openlan/pkg/libol"
	"github.com/luscis/openlan/pkg/schema"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type vpnClient struct {
	Directory string
}

func ParseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func (o *vpnClient) GetDevice(name string) string {
	sw := config.Manager.Switch
	if sw == nil {
		return ""
	}
	for _, n := range sw.Network {
		vpn := n.OpenVPN
		if vpn == nil {
			continue
		}
		if vpn.Network == name {
			return vpn.Device
		}
	}
	return ""
}

func (o *vpnClient) scanStatus(network string, reader io.Reader,
	clients map[string]*schema.VPNClient) error {
	readAt := "header"
	offset := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "OpenVPN CLIENT LIST" {
			readAt = "common"
			offset = 3
		}
		if line == "ROUTING TABLE" {
			readAt = "routing"
			offset = 2
		}
		if line == "GLOBAL STATS" {
			readAt = "global"
			offset = 1
		}
		if offset > 0 {
			offset -= 1
			continue
		}
		columns := strings.SplitN(line, ",", 5)
		switch readAt {
		case "common":
			if len(columns) == 5 {
				name := columns[0]
				remote := columns[1]
				client := &schema.VPNClient{
					Name:   name,
					Remote: remote,
					State:  "success",
					Device: o.GetDevice(network),
				}
				if rxc, err := ParseInt64(columns[2]); err == nil {
					client.RxBytes = rxc
				}
				if txc, err := ParseInt64(columns[3]); err == nil {
					client.TxBytes = txc
				}
				if len(columns[4]) > 0 {
					var uptime time.Time
					var err error
					if unicode.IsDigit(rune(columns[4][0])) {
						uptime, err = libol.GetLocalTime(libol.SimpleTime, columns[4])
					} else {
						uptime, err = libol.GetLocalTime(time.ANSIC, columns[4])
					}
					if err == nil {
						client.Uptime = uptime.Unix()
						client.AliveTime = time.Now().Unix() - client.Uptime
					} else {
						libol.Warn("vpnClient.scanStatus %s", err)
					}
				}
				clients[remote] = client
			}
		case "routing":
			if len(columns) == 4 {
				remote := columns[2]
				address := columns[0]
				if client, ok := clients[remote]; ok {
					client.Address = address
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (o *vpnClient) Dir(args ...string) string {
	values := append([]string{o.Directory}, args...)
	return filepath.Join(values...)
}

func (o *vpnClient) statusFile(name string) []string {
	files, err := filepath.Glob(o.Dir(name, "*server.status"))
	if err != nil {
		libol.Warn("vpnClient.statusFile %v", err)
	}
	return files
}

func (o *vpnClient) readStatus(network string) map[string]*schema.VPNClient {
	clients := make(map[string]*schema.VPNClient, 32)
	for _, file := range o.statusFile(network) {
		reader, err := os.Open(file)
		if err != nil {
			libol.Debug("vpnClient.readStatus %v", err)
			return nil
		}
		if err := o.scanStatus(network, reader, clients); err != nil {
			libol.Warn("vpnClient.readStatus %v", err)
		}
		reader.Close()
	}
	return clients
}

func (o *vpnClient) List(name string) <-chan *schema.VPNClient {
	c := make(chan *schema.VPNClient, 128)

	clients := o.readStatus(name)
	go func() {
		for _, v := range clients {
			c <- v
		}
		c <- nil //Finish channel by nil.
	}()

	return c
}

func (o *vpnClient) GetClientProfile(network, client, remote string) (string, error) {
	file := o.Dir(network, client+"client.ovpn")
	reader, err := os.Open(file)
	if err != nil {
		return "", err
	}
	profile := ""
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "remote 0.0.0.0") {
			profile += strings.Replace(line, "0.0.0.0", remote, 1)
		} else {
			profile += line
		}
		profile += "\n"
	}
	if err := scanner.Err(); err != nil {
		return profile, err
	}
	return profile, nil
}

var VPNClient = vpnClient{
	Directory: config.VarDir("openvpn"),
}
