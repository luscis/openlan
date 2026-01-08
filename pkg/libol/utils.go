package libol

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

const LeaseTime = "2006-01-02T15"
const SimpleTime = "2006-01-02 15:04:05"
const MacBase = 0x00

var Letters = []byte("0123456789abcdefghijklmnopqrstuvwxyz")

func IsYaml(file string) bool {
	return strings.HasSuffix(file, ".yaml")
}

func IsJson(file string) bool {
	return strings.HasSuffix(file, ".json")
}

func GenString(n int) string {
	buffer := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range buffer {
		buffer[i] = Letters[rand.Int63()%int64(len(Letters))]
	}
	buffer[0] = Letters[rand.Int63()%26+10]
	return string(buffer)
}

func GenLetters(n int) []byte {
	buffer := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range buffer {
		buffer[i] = Letters[rand.Int63()%int64(len(Letters))]
	}
	buffer[0] = Letters[rand.Int63()%26+10]
	return buffer
}

func GenEthAddr(n int) []byte {
	if n == 0 {
		n = 6
	}
	data := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range data {
		data[i] = byte(rand.Uint32() & 0xFF)
	}
	data[0] = MacBase
	return data
}

func GenUint32() uint32 {
	rand.Seed(time.Now().UnixNano())
	return rand.Uint32()
}

func GenInt32() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func Marshal(v interface{}, pretty bool) ([]byte, error) {
	str, err := json.Marshal(v)
	if err != nil {
		Error("Marshal error: %s", err)
		return nil, err
	}
	if !pretty {
		return str, nil
	}
	var out bytes.Buffer
	if err := json.Indent(&out, str, "", "  "); err != nil {
		return str, nil
	}
	return out.Bytes(), nil
}

func MarshalYaml(v interface{}) ([]byte, error) {
	str, err := yaml.Marshal(v)
	if err != nil {
		Error("Marshal error: %s", err)
		return nil, err
	}
	return str, nil
}

func MarshalSave(v interface{}, file string, pretty bool) error {
	f, err := CreateFile(file)
	if err != nil {
		Error("MarshalSave: %s", err)
		return err
	}
	defer f.Close()

	var data []byte
	if IsYaml(file) {
		data, err = MarshalYaml(v)
		if err != nil {
			Error("MarshalSave error: %s", err)
			return err
		}
	} else {
		data, err = Marshal(v, true)
		if err != nil {
			Error("MarshalSave error: %s", err)
			return err
		}
	}
	if _, err := f.Write(data); err != nil {
		Error("MarshalSave: %s", err)
		return err
	}
	return nil
}

func FileExist(file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return err
	}
	return nil
}

func LoadFile(file string) ([]byte, error) {
	return os.ReadFile(file)
}

func Unmarshal(v interface{}, contents []byte) error {
	if err := json.Unmarshal(contents, v); err != nil {
		return NewErr("%s", err)
	}
	return nil
}

func UnmarshalYaml(v interface{}, contents []byte) error {
	if err := yaml.Unmarshal(contents, v); err != nil {
		return NewErr("%s", err)
	}
	return nil
}

func UnmarshalLoad(v interface{}, file string) error {
	if err := FileExist(file); err != nil {
		return nil
	}
	contents, err := LoadFile(file)
	if err != nil {
		return NewErr("%s %s", file, err)
	}

	if IsYaml(file) {
		return UnmarshalYaml(v, contents)
	} else {
		return Unmarshal(v, contents)
	}
}

func FunName(i interface{}) string {
	ptr := reflect.ValueOf(i).Pointer()
	name := runtime.FuncForPC(ptr).Name()
	return path.Base(name)
}

func Netmask2Len(s string) int {
	mask := net.IPMask(net.ParseIP(s).To4())
	prefixSize, _ := mask.Size()
	return prefixSize
}

func IPNetmask(ipAddr string) (string, error) {
	if i, n, err := net.ParseCIDR(ipAddr); err == nil {
		return i.String() + "/" + net.IP(n.Mask).String(), nil
	} else {
		return "", err
	}
}

func IPNetwork(ipAddr string) (string, error) {
	if _, n, err := net.ParseCIDR(ipAddr); err == nil {
		return n.IP.String() + "/" + net.IP(n.Mask).String(), nil
	} else {
		return ipAddr, err
	}
}

func UnixTime(value int64) string {
	return time.Unix(value, 0).UTC().String()
}

func ToSegment(id int, sec string) string {
	if id != 0 {
		return fmt.Sprintf("ID:%d", id)
	}
	return fmt.Sprintf("USER:%s", strings.SplitN(sec, ":", 2)[0])
}

func PrettyTime(t int64) string {
	s := ""
	if t < 0 {
		s = "-"
		t = -t
	}
	min := t / 60
	if min < 60 {
		return fmt.Sprintf("%s%dm%ds", s, min, t%60)
	}
	hours := min / 60
	if hours < 24 {
		return fmt.Sprintf("%s%dh%dm", s, hours, min%60)
	}
	days := hours / 24
	return fmt.Sprintf("%s%dd%dh", s, days, hours%24)
}

func PrettyBytes(b uint64) string {
	b *= 8
	split := func(_v uint64, _m uint64) (i uint64, d int) {
		_d := float64(_v%_m) / float64(_m)
		return _v / _m, int(_d * 100) //move two decimal to integer
	}
	if b < 1024 {
		return fmt.Sprintf("%db", b)
	}
	k, d := split(b, 1024)
	if k < 1024 {
		return fmt.Sprintf("%d.%02dKb", k, d)
	}
	m, d := split(k, 1024)
	if m < 1024 {
		return fmt.Sprintf("%d.%02dMb", m, d)
	}
	g, d := split(m, 1024)
	return fmt.Sprintf("%d.%02dGb", g, d)
}

func GetIPAddr(addr string) string {
	_addr, _ := GetHostPort(addr)
	return _addr
}

func GetHostPort(addr string) (string, string) {
	values := strings.SplitN(addr, ":", 2)
	if len(values) == 2 {
		return values[0], values[1]
	}
	return values[0], ""
}

func Wait() {
	x := make(chan os.Signal, 1)
	signal.Notify(x, os.Interrupt, syscall.SIGTERM)
	signal.Notify(x, os.Interrupt, syscall.SIGQUIT) //CTL+/
	signal.Notify(x, os.Interrupt, syscall.SIGINT)  //CTL+C
	Info("Wait: ...")
	n := <-x
	Warn("Wait: ... Signal %d received ...", n)
}

func OpenTrunk(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
}

func OpenWrite(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
}

func OpenRead(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDONLY, os.ModePerm)
}

func CreateFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
}

func ParseAddr(addr string) net.IP {
	ip := strings.SplitN(addr, "/", 2)[0]
	return net.ParseIP(ip)
}

func GetPrefixLen(addr string) int {
	values := strings.SplitN(addr, "/", 2)
	if len(values) == 2 {
		size, _ := strconv.Atoi(values[1])
		return size
	}
	return 32
}

func ParseNet(addr string) (*net.IPNet, error) {
	return ParseCIDR(addr)
}

func ParseCIDR(addr string) (*net.IPNet, error) {
	if !strings.Contains(addr, "/") {
		addr = addr + "/32"
	}
	_, cidr, err := net.ParseCIDR(addr)
	return cidr, err
}

func Uint2S(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

func IfName(name string) string {
	size := len(name)
	if size < 15 {
		return name
	}
	return name[size-15 : size]
}

func GetLocalTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, time.Local)
}

func GetLeaseTime(value string) (time.Time, error) {
	return time.ParseInLocation(LeaseTime, value, time.Local)
}

func Base64Decode(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}

func Base64Encode(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func GetPrefix(value string, index int) string {
	if len(value) >= index {
		return value[:index]
	}
	return ""
}

func GetSuffix(value string, index int) string {
	if len(value) >= index {
		return value[index:]
	}
	return ""
}

func Sudo(bin string, args ...string) (string, error) {
	binArgs := append([]string{bin}, args...)
	out, err := exec.Command("sudo", binArgs...).CombinedOutput()
	return string(out), err
}

func Exec(bin string, args ...string) (string, error) {
	out, err := exec.Command(bin, args...).CombinedOutput()
	return string(out), err
}

func HasProcess(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func CreateFileEx(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
}

func WritePid(file string) {
	pid := os.Getpid()
	if fp, err := OpenTrunk(file); err == nil {
		fmt.Fprintf(fp, "%d\n", pid)
	}
}
