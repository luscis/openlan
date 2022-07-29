package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Package struct {
	Version        [2]byte // 协议版本
	Length         int16   // 数据部分长度
	Timestamp      int64   // 时间戳
	HostnameLength int16   // 主机名长度
	Hostname       []byte  // 主机名
	TagLength      int16   // Tag长度
	Tag            []byte  // Tag
	Msg            []byte  // 数据部分长度
}

func (p *Package) Pack(writer io.Writer) error {
	var err error
	err = binary.Write(writer, binary.BigEndian, &p.Version)
	err = binary.Write(writer, binary.BigEndian, &p.Length)
	err = binary.Write(writer, binary.BigEndian, &p.Timestamp)
	err = binary.Write(writer, binary.BigEndian, &p.HostnameLength)
	err = binary.Write(writer, binary.BigEndian, &p.Hostname)
	err = binary.Write(writer, binary.BigEndian, &p.TagLength)
	err = binary.Write(writer, binary.BigEndian, &p.Tag)
	err = binary.Write(writer, binary.BigEndian, &p.Msg)
	return err
}
func (p *Package) Unpack(reader io.Reader) error {
	var err error
	err = binary.Read(reader, binary.BigEndian, &p.Version)
	err = binary.Read(reader, binary.BigEndian, &p.Length)
	err = binary.Read(reader, binary.BigEndian, &p.Timestamp)
	err = binary.Read(reader, binary.BigEndian, &p.HostnameLength)
	p.Hostname = make([]byte, p.HostnameLength)
	err = binary.Read(reader, binary.BigEndian, &p.Hostname)
	err = binary.Read(reader, binary.BigEndian, &p.TagLength)
	p.Tag = make([]byte, p.TagLength)
	err = binary.Read(reader, binary.BigEndian, &p.Tag)
	p.Msg = make([]byte, p.Length-8-2-p.HostnameLength-2-p.TagLength)
	err = binary.Read(reader, binary.BigEndian, &p.Msg)
	return err
}

func (p *Package) String() string {
	return fmt.Sprintf("version:%s length:%d timestamp:%d hostname:%s tag:%s msg:%s",
		p.Version,
		p.Length,
		p.Timestamp,
		p.Hostname,
		p.Tag,
		p.Msg,
	)
}

func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF {
		return
	}

	log.Printf("INDEX: 0x%02x\n", data[0])
	if data[0] == 'V' {
		if len(data) > 4 {
			length := int16(0)
			binary.Read(bytes.NewReader(data[2:4]), binary.BigEndian, &length)
			if int(length)+4 <= len(data) {
				return int(length) + 4, data[:int(length)+4], nil
			}
		}
	}

	//scroll to next package.
	return 1, data[:1], nil
}

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	pack := &Package{
		Version:        [2]byte{'V', '1'},
		Timestamp:      time.Now().Unix(),
		HostnameLength: int16(len(hostname)),
		Hostname:       []byte(hostname),
		TagLength:      4,
		Tag:            []byte("demo"),
		Msg:            []byte(("现在时间是:" + time.Now().Format("2006-01-02 15:04:05"))),
	}
	pack.Length = 8 + 2 + pack.HostnameLength + 2 + pack.TagLength + int16(len(pack.Msg))

	buf := new(bytes.Buffer)
	// 写入四次，模拟TCP粘包效果
	pack.Pack(buf)
	pack.Pack(buf)
	pack.Pack(buf)
	pack.Pack(buf)
	buf.Write([]byte{0x00, 0x01, 0x02})
	pack.Pack(buf)

	buf.Write([]byte{'V', 0x01, 0x02, 0x11, 0x12})
	pack.Pack(buf)

	// scanner
	scanner := bufio.NewScanner(buf)
	scanner.Split(Split)
	for scanner.Scan() {
		scannedPack := new(Package)
		data := scanner.Bytes()
		if len(data) <= 1 {
			continue
		}
		scannedPack.Unpack(bytes.NewReader(data))
		log.Println(scannedPack)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("无效数据包 %s", err)
	}
}
