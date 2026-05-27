package libsock

import (
	"encoding/binary"
	"net"
	"strings"
	"testing"
	"time"
)

func TestFrameMessageControlDecode(t *testing.T) {
	f := NewControlFrame(NegoReq, []byte("hello"))
	if !f.IsControl() {
		t.Fatalf("expected control frame")
	}
	if ok := f.Decode(); !ok {
		t.Fatalf("decode should keep control=true")
	}
	action, params := f.CmdAndParams()
	if action != NegoReq {
		t.Fatalf("action mismatch: got=%q want=%q", action, NegoReq)
	}
	if string(params) != "hello" {
		t.Fatalf("params mismatch: got=%q want=%q", string(params), "hello")
	}
}

func TestFrameMessageAppendAndSetSize(t *testing.T) {
	f := NewFrameMessage(8)
	f.Append([]byte("abc"))
	if got := f.Size(); got != 3 {
		t.Fatalf("size mismatch after append: got=%d want=%d", got, 3)
	}
	f.SetSize(2)
	if got := f.Size(); got != 2 {
		t.Fatalf("size mismatch after setsize: got=%d want=%d", got, 2)
	}
	if string(f.Frame()[:f.Size()]) != "ab" {
		t.Fatalf("payload mismatch after setsize")
	}
}

func TestStreamDecodeReturnsNilOnPartialFrame(t *testing.T) {
	m := &StreamMessagerImpl{}
	full := make([]byte, HlSize+5)
	full[0] = MAGIC[0]
	full[1] = MAGIC[1]
	binary.BigEndian.PutUint16(full[HMSize:HlSize], 5)
	copy(full[HlSize:], []byte("hello"))

	frame, err := m.decode(full[:HlSize+2], 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if frame != nil {
		t.Fatalf("expected nil frame for partial packet")
	}
}

func TestStreamReceiveBufferFullReturnsError(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	m := &StreamMessagerImpl{
		timeout: time.Second,
		bufSize: 8,
	}

	go func() {
		// Header says payload 16, but receiver buffer is only 8.
		buf := make([]byte, HlSize+16)
		buf[0] = MAGIC[0]
		buf[1] = MAGIC[1]
		binary.BigEndian.PutUint16(buf[HMSize:HlSize], 16)
		copy(buf[HlSize:], []byte("0123456789abcdef"))
		_, _ = c1.Write(buf[:8]) // only fill receiver buffer once
		time.Sleep(20 * time.Millisecond)
		_ = c1.Close()
	}()

	_, err := m.Receive(c2, 1)
	if err == nil {
		t.Fatalf("expected error when receive buffer is full")
	}
	if !strings.Contains(err.Error(), "frame too large or incomplete") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEncodeMagicV1FrameLenIncludesNetID(t *testing.T) {
	payload := []byte("hello")
	network := "guest"
	buf := encodeMagicV1Frame(network, payload)

	if got, want := [2]byte{buf[0], buf[1]}, MAGICv1; got != want {
		t.Fatalf("magic mismatch: got=%x want=%x", got, want)
	}
	frameLen := int(binary.BigEndian.Uint16(buf[HMSize:HlSize]))
	if got, want := frameLen, V1Size+len(payload); got != want {
		t.Fatalf("len mismatch: got=%d want=%d", got, want)
	}
}

func TestDecodeFrameHeaderMagicV1(t *testing.T) {
	payload := []byte("world")
	network := "office"
	buf := encodeMagicV1Frame(network, payload)

	h, err := decodeFrameHeader(buf, 1)
	if err != nil {
		t.Fatalf("decode header failed: %v", err)
	}
	if h == nil {
		t.Fatalf("header should not be nil")
	}
	if h.magic != MAGICv1 {
		t.Fatalf("magic mismatch: got=%x want=%x", h.magic, MAGICv1)
	}
	if h.network != network {
		t.Fatalf("network mismatch: got=%q want=%q", h.network, network)
	}
	if h.payloadSize != V1Size+len(payload) {
		t.Fatalf("payload size mismatch: got=%d want=%d", h.payloadSize, V1Size+len(payload))
	}
	if h.frameLen != len(payload) {
		t.Fatalf("frameLen mismatch: got=%d want=%d", h.frameLen, len(payload))
	}
}

func TestDecodeFrameHeaderMagicLegacy(t *testing.T) {
	payload := []byte("legacy")
	buf := make([]byte, HlSize+len(payload))
	buf[0] = MAGIC[0]
	buf[1] = MAGIC[1]
	binary.BigEndian.PutUint16(buf[HMSize:HlSize], uint16(len(payload)))
	copy(buf[HlSize:], payload)

	h, err := decodeFrameHeader(buf, 1)
	if err != nil {
		t.Fatalf("decode header failed: %v", err)
	}
	if h == nil {
		t.Fatalf("header should not be nil")
	}
	if h.magic != MAGIC {
		t.Fatalf("magic mismatch: got=%x want=%x", h.magic, MAGIC)
	}
	if h.network != "" {
		t.Fatalf("network should be empty for legacy magic: got=%q", h.network)
	}
	if h.payloadSize != len(payload) {
		t.Fatalf("payload size mismatch: got=%d want=%d", h.payloadSize, len(payload))
	}
	if h.frameLen != len(payload) {
		t.Fatalf("frameLen mismatch: got=%d want=%d", h.frameLen, len(payload))
	}
}

func TestStreamSendMagicV1HeaderLayout(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	m := &StreamMessagerImpl{timeout: time.Second}
	frame := NewFrameMessage(16)
	frame.Append([]byte("abc"))
	setFrameMagic(frame, MAGICv1, "net-a")

	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, Hv1Size+frame.size)
		_, _ = c2.Read(buf)
		done <- buf
	}()

	if _, err := m.Send(c1, frame); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	buf := <-done
	if got, want := [2]byte{buf[0], buf[1]}, MAGICv1; got != want {
		t.Fatalf("magic mismatch: got=%x want=%x", got, want)
	}
	if got, want := int(binary.BigEndian.Uint16(buf[HMSize:HlSize])), V1Size+frame.size; got != want {
		t.Fatalf("len mismatch: got=%d want=%d", got, want)
	}
}

func TestStreamSendMagicLegacyHeaderLayout(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	m := &StreamMessagerImpl{timeout: time.Second}
	frame := NewFrameMessage(16)
	frame.Append([]byte("xyz"))

	done := make(chan []byte, 1)
	go func() {
		buf := make([]byte, HlSize+frame.size)
		_, _ = c2.Read(buf)
		done <- buf
	}()

	if _, err := m.Send(c1, frame); err != nil {
		t.Fatalf("send failed: %v", err)
	}

	buf := <-done
	if got, want := [2]byte{buf[0], buf[1]}, MAGIC; got != want {
		t.Fatalf("magic mismatch: got=%x want=%x", got, want)
	}
	if got, want := int(binary.BigEndian.Uint16(buf[HMSize:HlSize])), frame.size; got != want {
		t.Fatalf("len mismatch: got=%d want=%d", got, want)
	}
}

func TestStreamSendReceiveLegacyWithEncryption(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	key := "legacy-key"
	sender := &StreamMessagerImpl{timeout: time.Second}
	receiver := &StreamMessagerImpl{timeout: time.Second}
	sender.SetCrypt(NewBlockCrypt("xor", key))
	receiver.SetCrypt(NewBlockCrypt("xor", key))

	want := []byte("encrypted-legacy")
	frame := NewFrameMessage(len(want) + 8)
	frame.Append(want)

	errCh := make(chan error, 1)
	go func() {
		_, err := sender.Send(c1, frame)
		errCh <- err
	}()

	got, err := receiver.Receive(c2, 1)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if string(got.Frame()[:got.Size()]) != string(want) {
		t.Fatalf("payload mismatch: got=%q want=%q",
			string(got.Frame()[:got.Size()]), string(want))
	}
	if got.buffer[0] != MAGIC[0] || got.buffer[1] != MAGIC[1] {
		t.Fatalf("decoded legacy header magic mismatch: got=%x%x want=%x%x",
			got.buffer[0], got.buffer[1], MAGIC[0], MAGIC[1])
	}
}

func TestStreamSendReceiveMagicV1WithEncryption(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	oldResolver := ResolveNetworkCrypt
	defer func() { ResolveNetworkCrypt = oldResolver }()

	network := "team-a"
	key := "v1-network-key"
	ResolveNetworkCrypt = func(name string) *BlockCrypt {
		if name == network {
			return NewBlockCrypt("xor", key)
		}
		return nil
	}

	sender := &StreamMessagerImpl{timeout: time.Second}
	receiver := &StreamMessagerImpl{timeout: time.Second}
	sender.SetCrypt(NewBlockCrypt("xor", key))

	want := []byte("encrypted-v1")
	frame := NewFrameMessage(len(want) + 8)
	frame.Append(want)
	setFrameMagic(frame, MAGICv1, network)

	errCh := make(chan error, 1)
	go func() {
		_, err := sender.Send(c1, frame)
		errCh <- err
	}()

	got, err := receiver.Receive(c2, 1)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if string(got.Frame()[:got.Size()]) != string(want) {
		t.Fatalf("payload mismatch: got=%q want=%q",
			string(got.Frame()[:got.Size()]), string(want))
	}
	if got.buffer[0] != MAGICv1[0] || got.buffer[1] != MAGICv1[1] {
		t.Fatalf("decoded v1 payload should keep v1 frame header: got=%x%x want=%x%x",
			got.buffer[0], got.buffer[1], MAGICv1[0], MAGICv1[1])
	}
}

func TestPacketSendReceiveLegacyWithEncryption(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	key := "packet-legacy-key"
	sender := &PacketMessagerImpl{timeout: time.Second}
	receiver := &PacketMessagerImpl{timeout: time.Second}
	sender.SetCrypt(NewBlockCrypt("xor", key))
	receiver.SetCrypt(NewBlockCrypt("xor", key))

	want := []byte("packet-legacy")
	frame := NewFrameMessage(len(want) + 8)
	frame.Append(want)

	errCh := make(chan error, 1)
	go func() {
		_, err := sender.Send(c1, frame)
		errCh <- err
	}()

	got, err := receiver.Receive(c2, 1)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if string(got.Frame()[:got.Size()]) != string(want) {
		t.Fatalf("payload mismatch: got=%q want=%q",
			string(got.Frame()[:got.Size()]), string(want))
	}
	if got.buffer[0] != MAGIC[0] || got.buffer[1] != MAGIC[1] {
		t.Fatalf("decoded legacy header magic mismatch: got=%x%x want=%x%x",
			got.buffer[0], got.buffer[1], MAGIC[0], MAGIC[1])
	}
}

func TestPacketSendReceiveMagicV1WithEncryption(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	oldResolver := ResolveNetworkCrypt
	defer func() { ResolveNetworkCrypt = oldResolver }()

	network := "team-p"
	key := "packet-v1-key"
	ResolveNetworkCrypt = func(name string) *BlockCrypt {
		if name == network {
			return NewBlockCrypt("xor", key)
		}
		return nil
	}

	sender := &PacketMessagerImpl{timeout: time.Second}
	receiver := &PacketMessagerImpl{timeout: time.Second}
	sender.SetCrypt(NewBlockCrypt("xor", key))

	want := []byte("packet-v1")
	frame := NewFrameMessage(len(want) + 8)
	frame.Append(want)
	setFrameMagic(frame, MAGICv1, network)

	errCh := make(chan error, 1)
	go func() {
		_, err := sender.Send(c1, frame)
		errCh <- err
	}()

	got, err := receiver.Receive(c2, 1)
	if err != nil {
		t.Fatalf("receive failed: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if string(got.Frame()[:got.Size()]) != string(want) {
		t.Fatalf("payload mismatch: got=%q want=%q",
			string(got.Frame()[:got.Size()]), string(want))
	}
	if got.buffer[0] != MAGIC[0] || got.buffer[1] != MAGIC[1] {
		t.Fatalf("decoded v1 payload should be reconstructed as legacy frame header: got=%x%x want=%x%x",
			got.buffer[0], got.buffer[1], MAGIC[0], MAGIC[1])
	}
}

func TestPacketReceiveTruncatedFrameReturnsError(t *testing.T) {
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()

	receiver := &PacketMessagerImpl{timeout: time.Second}

	errCh := make(chan error, 1)
	go func() {
		buf := make([]byte, HlSize+2)
		buf[0] = MAGIC[0]
		buf[1] = MAGIC[1]
		binary.BigEndian.PutUint16(buf[HMSize:HlSize], 8)
		copy(buf[HlSize:], []byte("xx"))
		_, err := c1.Write(buf)
		errCh <- err
	}()

	_, err := receiver.Receive(c2, 1)
	if err == nil {
		t.Fatalf("expected truncated frame error")
	}
	if !strings.Contains(err.Error(), "truncated frame") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("writer failed: %v", err)
	}
}
