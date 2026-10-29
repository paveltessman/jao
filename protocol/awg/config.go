package awg

import (
	"crypto/ecdh"
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/paveltessman/jao/container"
	"github.com/paveltessman/jao/protocol"
)

const ImageTag = "jao/awg:latest"

// dataDir is the dir in the container.
const dataDir = "/opt/amnezia/awg"

const (
	confPath            = dataDir + "/awg0.conf"
	serverPublicKeyPath = dataDir + "/wireguard_server_public_key.key"
	pskPath             = dataDir + "/wireguard_psk.key"
)

// Config holds the parameters of an AmneziaWG 3.1 server.
// The defaults and the env names come from the Amnezia client,
// see client/core/installers/awgInstaller.cpp in amnezia-vpn/amnezia-client.
type Config struct {
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

func DefaultConfig() Config {
	return Config{
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

func (c Config) env() map[string]string {
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

		"SPECIAL_JUNK_1": c.SpecialJunk1,
	}
}

var ErrBadEnv = errors.New("awg: bad container env")

func ConfigFromEnv(name string, env map[string]string) (Config, error) {
	r := envReader{env: env}
	cfg := Config{
		Name:       name,
		Port:       r.port("AWG_SERVER_PORT"),
		SubnetIP:   r.str("AWG_SUBNET_IP"),
		SubnetCIDR: r.int("WIREGUARD_SUBNET_CIDR"),

		JunkPacketCount:   r.int("JUNK_PACKET_COUNT"),
		JunkPacketMinSize: r.int("JUNK_PACKET_MIN_SIZE"),
		JunkPacketMaxSize: r.int("JUNK_PACKET_MAX_SIZE"),

		InitPacketJunkSize:        r.int("INIT_PACKET_JUNK_SIZE"),
		ResponsePacketJunkSize:    r.int("RESPONSE_PACKET_JUNK_SIZE"),
		CookieReplyPacketJunkSize: r.int("COOKIE_REPLY_PACKET_JUNK_SIZE"),
		TransportPacketJunkSize:   r.int("TRANSPORT_PACKET_JUNK_SIZE"),

		InitPacketMagicHeader:      r.str("INIT_PACKET_MAGIC_HEADER"),
		ResponsePacketMagicHeader:  r.str("RESPONSE_PACKET_MAGIC_HEADER"),
		UnderloadPacketMagicHeader: r.str("UNDERLOAD_PACKET_MAGIC_HEADER"),
		TransportPacketMagicHeader: r.str("TRANSPORT_PACKET_MAGIC_HEADER"),

		HeaderProtectionKey:  r.str("HEADER_PROTECTION_KEY"),
		RekeyAfterTime:       r.str("REKEY_AFTER_TIME"),
		RekeyTimeout:         r.str("REKEY_TIMEOUT"),
		RejectAfterTime:      r.str("REJECT_AFTER_TIME"),
		KeepaliveTimeout:     r.str("KEEPALIVE_TIMEOUT"),
		MaxHandshakeAttempts: r.str("MAX_HANDSHAKE_ATTEMPTS"),
		RandomTrailers:       r.str("RANDOM_TRAILERS"),
		DisableCookies:       r.str("DISABLE_COOKIES"),

		SpecialJunk1: r.str("SPECIAL_JUNK_1"),
	}
	if r.err != nil {
		return Config{}, fmt.Errorf("%w in %s: %w", ErrBadEnv, name, r.err)
	}
	if cfg.SubnetIP == "" {
		return Config{}, fmt.Errorf("%w in %s: AWG_SUBNET_IP is empty", ErrBadEnv, name)
	}
	return cfg, nil
}

type envReader struct {
	env map[string]string
	err error
}

func (r *envReader) str(key string) string {
	return r.env[key]
}

func (r *envReader) int(key string) int {
	n, err := strconv.Atoi(r.env[key])
	if err != nil && r.err == nil {
		r.err = fmt.Errorf("%s: %w", key, err)
	}
	return n
}

func (r *envReader) port(key string) uint16 {
	n, err := strconv.ParseUint(r.env[key], 10, 16)
	if err != nil && r.err == nil {
		r.err = fmt.Errorf("%s: %w", key, err)
	}
	return uint16(n)
}

func (c Config) Spec(hostDataDir string) container.ContainerSpec {
	return container.ContainerSpec{
		Name:    c.Name,
		Image:   ImageTag,
		Labels:  map[string]string{protocol.Label: protocol.AWGLabel},
		Env:     c.env(),
		CapAdd:  []string{"NET_ADMIN", "SYS_MODULE"},
		Devices: []string{"/dev/net/tun"},
		Ports:   []container.PortBinding{{HostPort: c.Port, ContainerPort: c.Port, Protocol: container.UDP}},
		Mounts:  []container.Mount{{HostPath: hostDataDir, ContainerPath: dataDir}},
		Sysctls: map[string]string{
			"net.ipv4.ip_forward":              "1",
			"net.ipv4.conf.all.src_valid_mark": "1",
		},
		Restart: container.RestartUnlessStopped,
	}
}

func genKey() string {
	return base64.StdEncoding.EncodeToString(newPrivateKey())
}

func genKeyPair() (private, public string, err error) {
	key, err := ecdh.X25519().NewPrivateKey(newPrivateKey())
	if err != nil {
		return "", "", fmt.Errorf("awg: can't make a key pair: %w", err)
	}
	enc := base64.StdEncoding
	return enc.EncodeToString(key.Bytes()), enc.EncodeToString(key.PublicKey().Bytes()), nil
}

func newPrivateKey() []byte {
	key := make([]byte, 32)
	_, _ = cryptorand.Read(key)
	key[0] &= 248
	key[31] = (key[31] & 127) | 64
	return key
}

var ErrSubnetFull = errors.New("awg: no free IP in the subnet")

var allowedIPRe = regexp.MustCompile(`(?m)^\s*AllowedIPs\s*=\s*(\d+\.\d+\.\d+\.\d+)`)

func nextPeerIP(serverConf string, subnetIP string, subnetCIDR int) (netip.Addr, error) {
	addr, err := netip.ParseAddr(subnetIP)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("awg: bad subnet IP %q: %w", subnetIP, err)
	}
	subnet, err := addr.Prefix(subnetCIDR)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("awg: bad subnet %s/%d: %w", subnetIP, subnetCIDR, err)
	}

	used := map[netip.Addr]bool{}
	for _, m := range allowedIPRe.FindAllStringSubmatch(serverConf, -1) {
		if ip, err := netip.ParseAddr(m[1]); err == nil {
			used[ip] = true
		}
	}

	for ip := subnet.Addr().Next(); subnet.Contains(ip.Next()); ip = ip.Next() {
		if !used[ip] {
			return ip, nil
		}
	}
	return netip.Addr{}, fmt.Errorf("%w: %s", ErrSubnetFull, subnet)
}

const (
	clientDNS        = "1.1.1.1, 1.0.0.1"
	clientAllowedIPs = "0.0.0.0/0, ::/0"
	clientKeepalive  = "25-35"
)

// client holds the per-peer values of a client config.
type client struct {
	PrivateKey      string
	IP              netip.Addr
	ServerPublicKey string
	PresharedKey    string
}

// clientConfig makes the config that the user imports into their app.
func (c Config) clientConfig(cl client, endpointHost string) string {
	iface := [][2]string{
		{"Address", cl.IP.String() + "/32"},
		{"DNS", clientDNS},
		{"PrivateKey", cl.PrivateKey},
		{"Jc", strconv.Itoa(c.JunkPacketCount)},
		{"Jmin", strconv.Itoa(c.JunkPacketMinSize)},
		{"Jmax", strconv.Itoa(c.JunkPacketMaxSize)},
		{"S1", strconv.Itoa(c.InitPacketJunkSize)},
		{"S2", strconv.Itoa(c.ResponsePacketJunkSize)},
		{"S3", strconv.Itoa(c.CookieReplyPacketJunkSize)},
		{"S4", strconv.Itoa(c.TransportPacketJunkSize)},
		{"H1", c.InitPacketMagicHeader},
		{"H2", c.ResponsePacketMagicHeader},
		{"H3", c.UnderloadPacketMagicHeader},
		{"H4", c.TransportPacketMagicHeader},
		{"I1", c.SpecialJunk1},
		{"HeaderProtectionKey", c.HeaderProtectionKey},
		{"RekeyAfterTime", c.RekeyAfterTime},
		{"RekeyTimeout", c.RekeyTimeout},
		{"RejectAfterTime", c.RejectAfterTime},
		{"KeepaliveTimeout", c.KeepaliveTimeout},
		{"MaxHandshakeAttempts", c.MaxHandshakeAttempts},
		{"RandomTrailers", c.RandomTrailers},
		{"DisableCookies", c.DisableCookies},
	}
	peer := [][2]string{
		{"PublicKey", cl.ServerPublicKey},
		{"PresharedKey", cl.PresharedKey},
		{"AllowedIPs", clientAllowedIPs},
		{"Endpoint", net.JoinHostPort(endpointHost, strconv.Itoa(int(c.Port)))},
		{"PersistentKeepalive", clientKeepalive},
	}

	var b strings.Builder
	writeSection(&b, "Interface", iface)
	b.WriteString("\n")
	writeSection(&b, "Peer", peer)
	return b.String()
}

func writeSection(b *strings.Builder, name string, lines [][2]string) {
	fmt.Fprintf(b, "[%s]\n", name)
	for _, l := range lines {
		if l[1] != "" {
			fmt.Fprintf(b, "%s = %s\n", l[0], l[1])
		}
	}
}
