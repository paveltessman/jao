package installer

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"math/rand/v2"
	"strconv"
)

const awgImageTag = "jao/awg:latest"

// awgDataDir is the dir in the container with the server keys and awg0.conf.
const awgDataDir = "/opt/amnezia/awg"

// AwgConfig holds the parameters of an AmneziaWG 3.1 server.
// The defaults and the env names come from the Amnezia client,
// see client/core/installers/awgInstaller.cpp in amnezia-vpn/amnezia-client.
type AwgConfig struct {
	Name       string
	Port       uint16
	SubnetIP   string
	SubnetCIDR int

	JunkPacketCount   int
	JunkPacketMinSize int
	JunkPacketMaxSize int

	InitPacketJunkSize        int
	ResponsePacketJunkSize    int
	CookieReplyPacketJunkSize int
	TransportPacketJunkSize   int

	InitPacketMagicHeader      string
	ResponsePacketMagicHeader  string
	UnderloadPacketMagicHeader string
	TransportPacketMagicHeader string

	HeaderProtectionKey  string
	RekeyAfterTime       string
	RekeyTimeout         string
	RejectAfterTime      string
	KeepaliveTimeout     string
	MaxHandshakeAttempts string
	RandomTrailers       string
	DisableCookies       string

	// SpecialJunk1 goes to the client config only.
	SpecialJunk1 string
}

// DefaultAwgConfig returns the Amnezia defaults with a random junk packet count
// and a new header protection key.
func DefaultAwgConfig() AwgConfig {
	return AwgConfig{
		Name:       "jao-awg",
		Port:       55424,
		SubnetIP:   "10.8.1.0",
		SubnetCIDR: 24,

		JunkPacketCount:   4 + rand.IntN(3),
		JunkPacketMinSize: 10,
		JunkPacketMaxSize: 50,

		InitPacketJunkSize:        12,
		ResponsePacketJunkSize:    12,
		CookieReplyPacketJunkSize: 12,
		TransportPacketJunkSize:   12,

		InitPacketMagicHeader:      "1",
		ResponsePacketMagicHeader:  "2",
		UnderloadPacketMagicHeader: "3",
		TransportPacketMagicHeader: "4",

		HeaderProtectionKey:  genKey(),
		RekeyAfterTime:       "100-120",
		RekeyTimeout:         "3-7",
		RejectAfterTime:      "150-180",
		KeepaliveTimeout:     "5-15",
		MaxHandshakeAttempts: "15-20",
		RandomTrailers:       "on",
		DisableCookies:       "on",

		SpecialJunk1: "<r 2><b 0x858000010001000000000669636c6f756403636f6d0000010001c00c000100010000105a00044d583737>",
	}
}

func (c AwgConfig) env() map[string]string {
	return map[string]string{
		"AWG_SUBNET_IP":         c.SubnetIP,
		"WIREGUARD_SUBNET_CIDR": strconv.Itoa(c.SubnetCIDR),
		"AWG_SERVER_PORT":       strconv.Itoa(int(c.Port)),

		"JUNK_PACKET_COUNT":    strconv.Itoa(c.JunkPacketCount),
		"JUNK_PACKET_MIN_SIZE": strconv.Itoa(c.JunkPacketMinSize),
		"JUNK_PACKET_MAX_SIZE": strconv.Itoa(c.JunkPacketMaxSize),

		"INIT_PACKET_JUNK_SIZE":         strconv.Itoa(c.InitPacketJunkSize),
		"RESPONSE_PACKET_JUNK_SIZE":     strconv.Itoa(c.ResponsePacketJunkSize),
		"COOKIE_REPLY_PACKET_JUNK_SIZE": strconv.Itoa(c.CookieReplyPacketJunkSize),
		"TRANSPORT_PACKET_JUNK_SIZE":    strconv.Itoa(c.TransportPacketJunkSize),

		"INIT_PACKET_MAGIC_HEADER":      c.InitPacketMagicHeader,
		"RESPONSE_PACKET_MAGIC_HEADER":  c.ResponsePacketMagicHeader,
		"UNDERLOAD_PACKET_MAGIC_HEADER": c.UnderloadPacketMagicHeader,
		"TRANSPORT_PACKET_MAGIC_HEADER": c.TransportPacketMagicHeader,

		"HEADER_PROTECTION_KEY":  c.HeaderProtectionKey,
		"REKEY_AFTER_TIME":       c.RekeyAfterTime,
		"REKEY_TIMEOUT":          c.RekeyTimeout,
		"REJECT_AFTER_TIME":      c.RejectAfterTime,
		"KEEPALIVE_TIMEOUT":      c.KeepaliveTimeout,
		"MAX_HANDSHAKE_ATTEMPTS": c.MaxHandshakeAttempts,
		"RANDOM_TRAILERS":        c.RandomTrailers,
		"DISABLE_COOKIES":        c.DisableCookies,
	}
}

func (c AwgConfig) spec(hostDataDir string) ContainerSpec {
	return ContainerSpec{
		Name:    c.Name,
		Image:   awgImageTag,
		Env:     c.env(),
		CapAdd:  []string{"NET_ADMIN", "SYS_MODULE"},
		Devices: []string{"/dev/net/tun"},
		Ports:   []PortBinding{{HostPort: c.Port, ContainerPort: c.Port, Protocol: UDP}},
		Mounts:  []Mount{{Source: hostDataDir, Target: awgDataDir}},
		Sysctls: map[string]string{
			"net.ipv4.ip_forward":              "1",
			"net.ipv4.conf.all.src_valid_mark": "1",
		},
		Restart: RestartUnlessStopped,
	}
}

// genKey makes a key in the same format as `awg genkey`:
// a clamped Curve25519 private key in base64.
func genKey() string {
	key := make([]byte, 32)
	_, _ = cryptorand.Read(key)
	key[0] &= 248
	key[31] = (key[31] & 127) | 64
	return base64.StdEncoding.EncodeToString(key)
}
