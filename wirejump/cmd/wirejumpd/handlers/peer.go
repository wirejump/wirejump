package handlers

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"wirejump/internal/ipc"
	"wirejump/internal/network"
	"wirejump/internal/state"
	"wirejump/internal/utils"
)

// Add downstream peer
func AddPeer(State *state.AppState, Pubkey string, Isolated bool) (string, error) {
	if !network.IsValidKey(Pubkey) {
		return "", errors.New("invalid public key")
	}

	pool := []netip.Addr{}
	conf, err := State.Network.Downstream.ReadConfig()

	if err != nil {
		return "", err
	}

	// Iterate all peers
	for _, peer := range conf["Peer"] {
		ips := strings.Trim(peer["AllowedIPs"], " ")
		key := strings.Trim(peer["PublicKey"], " ")

		// Check if key is already present
		if key == Pubkey {
			return "", errors.New("peer with this key is already registered")
		}

		// Contains server network CIDR
		if strings.Contains(ips, ",") {
			parts := strings.Split(ips, ",")

			// Only single IP required
			for _, part := range parts {
				prefix, err := netip.ParsePrefix(strings.Trim(part, " "))

				if err != nil {
					return "", err
				}

				if prefix.IsSingleIP() {
					pool = append(pool, prefix.Addr())
				}
			}
		} else {
			addr, err := netip.ParsePrefix(ips)

			if err != nil {
				return "", err
			}

			pool = append(pool, addr.Addr())
		}
	}

	// Get interface prefix
	prefix, err := netip.ParsePrefix(strings.Trim(conf["Interface"][0]["Address"], " "))

	if err != nil {
		return "", err
	}

	if prefix.Bits() == -1 {
		return "", errors.New("downstream interface has invalid network prefix")
	}

	free, err := network.GetFreeIP(pool, prefix)

	if err != nil {
		return "", err
	}

	// Format peer IP address. For the server config file,
	// /32 netmask is mandatory since it indicates main peer
	// address. If peer's not isolated, whole downstream network
	// will be appended along with the prefix. For the client
	// configuration, however, it doesn't make much sense because
	// client expects interface address along with that network prefix.
	// WireGuard will control which IPs can be routed for that particular
	// peer, so there's no harm if a peer tries to communicate outside
	// it's AllowedIPs range – it will receive a routing error.
	ipv4 := free.String()
	addr := []string{fmt.Sprintf("%s/%d", ipv4, 32)}

	// Create peer
	peer := utils.INIPair{
		"PublicKey": Pubkey,
	}

	// If peer is not isolated (default), add interface network to allowed IPs
	if !Isolated {
		addr = append(addr, prefix.Masked().String())
	}

	peer["AllowedIPs"] = strings.Join(addr, ", ")

	// Update config
	conf["Peer"] = append(conf["Peer"], peer)

	// Write new config
	if err := State.Network.Downstream.WriteConfig(conf); err != nil {
		return "", err
	}

	// Update interface state for this peer
	if err := State.Network.Downstream.UpdatePeerConfig("add", Pubkey, addr); err != nil {
		return "", err
	}

	// Client needs address with network prefix
	formatted := fmt.Sprintf("%s/%d", ipv4, prefix.Bits())

	return formatted, nil
}

// Remove downstream peer
func RemovePeer(State *state.AppState, Pubkey string) error {
	conf, err := State.Network.Downstream.ReadConfig()

	if err != nil {
		return err
	}

	// Desired peer index
	index := -1

	// Iterate all peers
	for i, peer := range conf["Peer"] {
		if peer["PublicKey"] == Pubkey {
			index = i

			break
		}
	}

	if index == -1 {
		return errors.New("peer with this key is not found")
	}

	// Remove peer; not the fastest method, but should be quick enough
	peers := append([]utils.INIPair{}, conf["Peer"][:index]...)
	conf["Peer"] = append(peers, conf["Peer"][index+1:]...)

	// Write config back
	if err := State.Network.Downstream.WriteConfig(conf); err != nil {
		return err
	}

	// Update interface state for this peer
	if err := State.Network.Downstream.UpdatePeerConfig("remove", Pubkey, []string{}); err != nil {
		return err
	}

	return nil
}

// Add or remove downstream peers
func (h *IpcHandler) ManagePeers(State *state.AppState, Params *ipc.PeerCommandRequest, Reply *interface{}) error {
	switch Params.Operation {
	case ipc.PeerCommandAddPeer:
		ipv4, err := AddPeer(State, Params.Pubkey, Params.Isolated)

		if err != nil {
			return err
		}

		reply := ipc.PeerCommandReply{}
		reply.Peer.Isolated = Params.Isolated
		reply.Peer.IPv4Address = ipv4

		*Reply = reply

		return nil
	case ipc.PeerCommandDeletePeer:
		return RemovePeer(State, Params.Pubkey)
	default:
		return fmt.Errorf("unknown peer operation: %d", Params.Operation)
	}
}
