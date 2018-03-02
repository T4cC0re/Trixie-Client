package Trixie

type Config struct {
	TrixieURL         string
	AuthToken         string
	AuthTokenValidity uint32
}

const (
	DefaultURL = "https://trixie.bigpoint.net"
)
