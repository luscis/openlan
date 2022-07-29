package network

type Framer struct {
	Data   []byte
	Source Taper
	Output Taper
}
