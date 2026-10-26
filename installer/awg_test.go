package installer

import (
	"encoding/base64"
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
