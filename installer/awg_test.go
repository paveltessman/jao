package installer

import (
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"net/netip"
	"strconv"
	"strings"
	"testing"
)

func TestDefaultAwgConfigUsesAmneziaValues(t *testing.T) {

	cfg := DefaultAwgConfig()

	want := map[string]string{
		"JUNK_PACKET_MIN_SIZE":          "10",
		"JUNK_PACKET_MAX_SIZE":          "50",
		"INIT_PACKET_JUNK_SIZE":         "12",
		"RESPONSE_PACKET_JUNK_SIZE":     "12",
		"COOKIE_REPLY_PACKET_JUNK_SIZE": "12",
		"TRANSPORT_PACKET_JUNK_SIZE":    "12",
		"INIT_PACKET_MAGIC_HEADER":      "1",
		"RESPONSE_PACKET_MAGIC_HEADER":  "2",
		"UNDERLOAD_PACKET_MAGIC_HEADER": "3",
		"TRANSPORT_PACKET_MAGIC_HEADER": "4",
		"REKEY_AFTER_TIME":              "100-120",
		"REKEY_TIMEOUT":                 "3-7",
		"REJECT_AFTER_TIME":             "150-180",
		"KEEPALIVE_TIMEOUT":             "5-15",
		"MAX_HANDSHAKE_ATTEMPTS":        "15-20",
		"RANDOM_TRAILERS":               "on",
		"DISABLE_COOKIES":               "on",
		"AWG_SERVER_PORT":               "55424",
	}
	env := cfg.env()
	for key, value := range want {
		if env[key] != value {
			t.Errorf("env %s: want %q, got %q", key, value, env[key])
		}
	}
	for key, value := range env {
		if value == "" {
			t.Errorf("env %s: want a value, got empty string", key)
		}
	}
}

func TestDefaultAwgConfigJunkPacketCountInRange(t *testing.T) {

	for range 200 {
		jc := DefaultAwgConfig().JunkPacketCount
		if jc < 4 || jc > 6 {
			t.Fatalf("junk packet count: want 4..6, got %d", jc)
		}
	}
}

func TestDefaultAwgConfigMakesNewHeaderProtectionKey(t *testing.T) {

	first := DefaultAwgConfig().HeaderProtectionKey
	second := DefaultAwgConfig().HeaderProtectionKey

	if first == second {
		t.Errorf("want a new key for each config, got %s twice", first)
	}

	key, err := base64.StdEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("decode key: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("key length: want 32, got %d", len(key))
	}
	if key[0]&7 != 0 || key[31]&128 != 0 || key[31]&64 == 0 {
		t.Errorf("want a clamped Curve25519 key, got %x", key)
	}
}

func TestGenKeyPairPublicKeyMatchesPrivateKey(t *testing.T) {

	private, public, err := genKeyPair()
	if err != nil {
		t.Fatalf("gen key pair: %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(private)
	if err != nil {
		t.Fatalf("decode private key: %v", err)
	}
	key, err := ecdh.X25519().NewPrivateKey(raw)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	want := base64.StdEncoding.EncodeToString(key.PublicKey().Bytes())
	if public != want {
		t.Errorf("public key: want %s, got %s", want, public)
	}
	if raw[0]&7 != 0 || raw[31]&128 != 0 || raw[31]&64 == 0 {
		t.Errorf("want a clamped Curve25519 key, got %x", raw)
	}

	_, other, err := genKeyPair()
	if err != nil {
		t.Fatalf("gen key pair: %v", err)
	}
	if other == public {
		t.Errorf("want a new key pair for each call, got %s twice", public)
	}
}

const serverConfWithPeers = `[Interface]
PrivateKey = c2VydmVy
Address = 10.8.1.0/24
ListenPort = 55424

[Peer]
PublicKey = YQ==
PresharedKey = cHNr
AllowedIPs = 10.8.1.1/32

[Peer]
PublicKey = Yg==
PresharedKey = cHNr
AllowedIPs = 10.8.1.3/32
`

func TestNextPeerIP(t *testing.T) {

	tests := []struct {
		name string
		conf string
		cidr int
		want string
	}{
		{"no peers skips the server address", "[Interface]\nAddress = 10.8.1.0/24\n", 24, "10.8.1.1"},
		{"fills a gap", serverConfWithPeers, 24, "10.8.1.2"},
		{"ignores the server address line", "Address = 10.8.1.1/24\n", 24, "10.8.1.1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextPeerIP(tt.conf, "10.8.1.0", tt.cidr)
			if err != nil {
				t.Fatalf("next peer IP: %v", err)
			}
			if got.String() != tt.want {
				t.Errorf("want %s, got %s", tt.want, got)
			}
		})
	}
}

func TestNextPeerIPSkipsBroadcast(t *testing.T) {

	// 10.8.1.0/30: .0 is the server, .3 is the broadcast address.
	conf := "AllowedIPs = 10.8.1.1/32\n"

	got, err := nextPeerIP(conf, "10.8.1.0", 30)
	if err != nil {
		t.Fatalf("next peer IP: %v", err)
	}
	if got != netip.MustParseAddr("10.8.1.2") {
		t.Errorf("want 10.8.1.2, got %s", got)
	}

	conf += "AllowedIPs = 10.8.1.2/32\n"
	_, err = nextPeerIP(conf, "10.8.1.0", 30)
	if !errors.Is(err, ErrSubnetFull) {
		t.Errorf("want ErrSubnetFull, got %v", err)
	}
}

func TestNextPeerIPRejectsBadSubnet(t *testing.T) {

	if _, err := nextPeerIP("", "10.8.1", 24); err == nil {
		t.Error("bad IP: want an error, got nil")
	}
	if _, err := nextPeerIP("", "10.8.1.0", 33); err == nil {
		t.Error("bad CIDR: want an error, got nil")
	}
}

func TestClientConfigHoldsServerParams(t *testing.T) {

	cfg := DefaultAwgConfig()
	cfg.Port = 40001
	cl := awgClient{
		PrivateKey:      "Y2xpZW50",
		IP:              netip.MustParseAddr("10.8.1.2"),
		ServerPublicKey: "c2VydmVy",
		PresharedKey:    "cHNr",
	}

	conf := cfg.clientConfig(cl, "203.0.113.7")

	for _, want := range []string{
		"[Interface]\nAddress = 10.8.1.2/32\n",
		"DNS = 1.1.1.1, 1.0.0.1\n",
		"PrivateKey = Y2xpZW50\n",
		"Jc = " + strconv.Itoa(cfg.JunkPacketCount) + "\n",
		"Jmin = 10\n",
		"Jmax = 50\n",
		"S1 = 12\n",
		"S4 = 12\n",
		"H1 = 1\n",
		"H4 = 4\n",
		"I1 = " + cfg.SpecialJunk1 + "\n",
		"HeaderProtectionKey = " + cfg.HeaderProtectionKey + "\n",
		"RekeyAfterTime = 100-120\n",
		"DisableCookies = on\n",
		"\n[Peer]\nPublicKey = c2VydmVy\n",
		"PresharedKey = cHNr\n",
		"AllowedIPs = 0.0.0.0/0, ::/0\n",
		"Endpoint = 203.0.113.7:40001\n",
		"PersistentKeepalive = 25-35\n",
	} {
		if !strings.Contains(conf, want) {
			t.Errorf("want %q in config, got:\n%s", want, conf)
		}
	}
}

func TestClientConfigDropsEmptyValues(t *testing.T) {

	cfg := DefaultAwgConfig()
	cfg.SpecialJunk1 = ""
	cfg.RandomTrailers = ""
	cl := awgClient{IP: netip.MustParseAddr("10.8.1.2")}

	conf := cfg.clientConfig(cl, "2001:db8::1")

	for _, line := range strings.Split(conf, "\n") {
		if strings.HasSuffix(line, " =") || strings.HasSuffix(line, " = ") {
			t.Errorf("want no empty values, got line %q", line)
		}
	}
	if strings.Contains(conf, "I1 =") || strings.Contains(conf, "RandomTrailers =") {
		t.Errorf("want I1 and RandomTrailers dropped, got:\n%s", conf)
	}
	if !strings.Contains(conf, "Endpoint = [2001:db8::1]:55424\n") {
		t.Errorf("want a bracketed IPv6 endpoint, got:\n%s", conf)
	}
}
