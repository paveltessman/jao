package awg

import (
	"context"
	"fmt"
	"strings"
)

// Execer runs a command in a container and returns its stdout.
type Execer interface {
	ContainerExec(ctx context.Context, id string, cmd []string) (string, error)
}

// addPeerScript appends the peer in $1 to awg0.conf and applies the file
// to awg0 without a restart. The peer comes as an argument, not as part of
// the script, so bash does not read it.
const addPeerScript = `printf '%s' "$1" >> ` + confPath + ` && awg syncconf awg0 <(awg-quick strip ` + confPath + `)`

// AddPeer adds a new peer to the AWG container from cfg and returns
// the client config for it. endpointHost is the public address of the node,
// and it must not be empty.
// The client private key is not kept on the server.
// The caller must make sure that two calls for one container do not run
// at the same time, because the two peers can get the same IP.
func AddPeer(ctx context.Context, ctr Execer, cfg Config, endpointHost string) (string, error) {
	serverConf, err := readFile(ctx, ctr, cfg.Name, confPath)
	if err != nil {
		return "", err
	}
	serverPublicKey, err := readFile(ctx, ctr, cfg.Name, serverPublicKeyPath)
	if err != nil {
		return "", err
	}
	psk, err := readFile(ctx, ctr, cfg.Name, pskPath)
	if err != nil {
		return "", err
	}

	ip, err := nextPeerIP(serverConf, cfg.SubnetIP, cfg.SubnetCIDR)
	if err != nil {
		return "", err
	}
	privateKey, publicKey, err := genKeyPair()
	if err != nil {
		return "", err
	}

	peer := fmt.Sprintf("\n[Peer]\nPublicKey = %s\nPresharedKey = %s\nAllowedIPs = %s/32\n", publicKey, psk, ip)
	_, err = ctr.ContainerExec(ctx, cfg.Name, []string{"bash", "-c", addPeerScript, "_", peer})
	if err != nil {
		return "", fmt.Errorf("awg: can't add peer %s to %s: %w", ip, cfg.Name, err)
	}

	cl := client{
		PrivateKey:      privateKey,
		IP:              ip,
		ServerPublicKey: serverPublicKey,
		PresharedKey:    psk,
	}
	return cfg.clientConfig(cl, endpointHost), nil
}

// readFile returns the content of a file in the container without
// the spaces at the ends.
func readFile(ctx context.Context, ctr Execer, ctrName, path string) (string, error) {
	out, err := ctr.ContainerExec(ctx, ctrName, []string{"cat", path})
	if err != nil {
		return "", fmt.Errorf("awg: can't read %s in %s: %w", path, ctrName, err)
	}
	return strings.TrimSpace(out), nil
}
