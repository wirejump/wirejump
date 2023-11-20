package providers

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Conceptually there are 3 pillars of Mullvad API:
// - API auth token, used for protected API
// - account/device API, protected
// - relays (servers), unprotected
//
// First two are connected (devices and accounts are private), and relays can
// be queued without authentication. While it seems there's no public documentation
// available for auth API, opensource nature of Mullvad makes it quite easy to find.
//
// Some reference code: https://github.com/mullvad/mullvadvpn-app/blob/main/mullvad-api/src
//
//
// As of 2023 Mullvad is using new (compared to 2021) accounts->devices API,
// in which each account can have up to 5 devices associated with it.
// Each virtual device has unique name and id and can serve single connection,
// so each account supports up to 5 distinct connections total.
//
// When using WireGuard, each device holds a single pubkey. API has support for
// pubkey rotation, so there's no need to recreate a device each time (though
// it's probably better for the privacy). However, if single account is shared
// between WireJump server and another machine, key management can become tricky.
//
// API reference: https://api.mullvad.net/accounts/v1/
//
//
// There's also new relays API, and it looks like old www one is being deprecated, since
// web-based server selector does not make any visible API requests. This new relays API
// is being used by Mullvad app, so it's safe to assume it's not going anywhere.
// However, the API structure has been reworked greatly. Now, each server needs two
// additional info objects: WireGuard port range and Location info.
//
// API reference: https://api.mullvad.net/app/documentation
//
// Presented here data structures are incomplete in purpose,
// containing only required fields for this provider. Old API is still up but have different
// structure and is available at: https://api.mullvad.net/public/relays/wireguard/v1/
//

// API URL
var mullvadAPIBaseURL = "https://api.mullvad.net"

// Mullvad uses DNS hijacking, which catches all DNS requests coming
// from the tunnel and redirects them to their DNS servers to prevent
// DNS leaks. Since WireJump has unbound installed on the server, and
// unbound does not use this tunnel, setting this to true makes sense
// if network users specify their own DNS (hardcoding 1.1.1.1, for example).
// In this case, Mullvad will catch it and prevent DNS leak. This should
// probably exist as a separate setting.
const HijackDNSOption = true

// Holds auth data
var authToken *mullvadAuthToken

// Represents error returned by API
type mullvadAPIError struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details"`
}

// Represents auth token structure
type mullvadAuthToken struct {
	Token  string `json:"access_token"`
	Expiry string `json:"expiry"`
}

// Auth token request
type mullvadAuthTokenRequest struct {
	Account string `json:"account_number"`
}

// Detailed server location
type mullvadServerLocation struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

// Describes one of the relays
type mullvadWireguardServer struct {
	Hostname string `json:"hostname"`
	Location string `json:"location"`
	Active   bool   `json:"active"`
	Owned    bool   `json:"owned"`
	IPv4Addr string `json:"ipv4_addr_in"`
	Pubkey   string `json:"public_key"`
}

// Wireguard server wrapper with some additional info
type mullvadWireguardServerWrapper struct {
	PortRanges [][]int                  `json:"port_ranges"`
	Relays     []mullvadWireguardServer `json:"relays"`
}

// List of WireGuard servers with all required connection info
type mullvadServersList struct {
	Locations map[string]mullvadServerLocation `json:"locations"`
	Wireguard mullvadWireguardServerWrapper    `json:"wireguard"`
}

// Describes Mullvad account
type mullvadAccount struct {
	Id            string `json:"string"`
	Expiry        string `json:"expiry"`
	MaxDevices    int    `json:"max_devices"`
	CanAddDevices bool   `json:"can_add_devices"`
	Number        int64  `json:"number"`
}

// Describes Mullvad device
type mullvadDevice struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Pubkey      string `json:"pubkey"`
	HijackDNS   bool   `json:"hijack_dns"`
	Created     string `json:"created"`
	IPv4Address string `json:"ipv4_address"`
	IPv6Address string `json:"ipv6_address"`
}

// Used for device creation & pubkey rotation
type mullvadDeviceRequest struct {
	Pubkey    string `json:"pubkey"`
	HijackDNS bool   `json:"hijack_dns,omitempty"`
}

// Parse Mullvad-specific expiration time format
func parseExpiry(value string) (time.Time, error) {
	// Go time format is REALLY weird...
	expiry, err := time.Parse("2006-01-02T15:04:05-07:00", value)

	if err != nil {
		return time.Time{}, err
	}

	return expiry, nil
}

// Check auth token for validity
func tokenIsValid() bool {
	if authToken != nil {
		// Check expiration date...
		if authToken.Expiry != "" {
			expiry, err := parseExpiry(authToken.Expiry)

			if err != nil {
				return false
			}

			// Token is still active
			if time.Now().Before(expiry) {
				return true
			}
		}
	}

	return false
}

// Mullvad-specific API request. Will update auth token automatically if needed
func (m *WireguardProvider) APIRequest(Method string, URL string, UseAuth bool, Data interface{}, Reply interface{}) error {
	headers := make(http.Header)
	api_error := mullvadAPIError{}

	// Check if token is required
	if UseAuth {
		// Token needs to be refreshed
		if !tokenIsValid() {
			url := m.URL("auth", "v1", "token")
			req := mullvadAuthTokenRequest{Account: m.Account.AccountID}
			token := mullvadAuthToken{}

			failed, err := RequestAPI("POST", url, nil, req, &token, &api_error)

			if err != nil {
				if failed {
					return fmt.Errorf("failed to refresh token [%s]: %s", api_error.Code, api_error.Details)
				}

				return err
			}

			authToken = &token
		}

		// Create auth header
		headers.Add("Authorization", fmt.Sprintf("Bearer %s", authToken.Token))
	}

	// Make API request
	api_failed, err := RequestAPI(Method, URL, headers, Data, Reply, &api_error)

	if err != nil {
		if api_failed {
			return fmt.Errorf("API error [%s]: %s", api_error.Code, api_error.Details)
		}

		return err
	}

	return nil
}

// This will be run on startup to register Mullvad provider
func MullvadInit(Account WireguardProviderAccount) (p WireguardProvider, e error) {
	if Account.AccountID == "" {
		e = errors.New("AccountID as string is required for Mullvad")
	} else {
		e = nil
		p = WireguardProvider{
			Account:      Account,
			Initialized:  true,
			ProviderName: "mullvad",

			// This address seems to be static across the years, although
			// new API returns upstream gateway address explicitly now
			UpstreamGateway: "10.64.0.1",
			URL:             FormatURL(mullvadAPIBaseURL, false),
		}

		// initialize token
		authToken = nil
	}

	return p, e
}

// Fetch account info
func (m *WireguardProvider) GetAccountInfo() (WireguardAccount, error) {
	acc := mullvadAccount{}
	url := m.URL("accounts", "v1", "accounts", "me")
	err := m.APIRequest("GET", url, true, nil, &acc)

	if err != nil {
		return WireguardAccount{}, err
	}

	expiry, err := parseExpiry(acc.Expiry)

	if err != nil {
		return WireguardAccount{}, err
	}

	if !acc.CanAddDevices {
		return WireguardAccount{}, errors.New("this account can not add new devices")
	}

	return WireguardAccount{
		Expires: expiry.Unix(),
	}, nil
}

// Get all active & owned (as claimed by Mullvad) WireGuard servers
func (m *WireguardProvider) GetAllServers() ([]WireguardServer, error) {
	all_servers := []WireguardServer{}
	mullvadObject := mullvadServersList{}
	url := m.URL("app", "v1", "relays")
	err := m.APIRequest("GET", url, false, nil, &mullvadObject)

	if err != nil {
		return []WireguardServer{}, err
	}

	// Process port ranges first;
	// add them all to a single pool to randomly query from later
	var incomingPorts []int

	for _, portRange := range mullvadObject.Wireguard.PortRanges {
		for port := portRange[0]; port <= portRange[1]; port++ {
			incomingPorts = append(incomingPorts, port)
		}
	}

	// Assemble final list
	for _, relay := range mullvadObject.Wireguard.Relays {
		if relay.Active && relay.Owned {
			port := GetRandomElement(incomingPorts)

			if port == nil {
				return []WireguardServer{}, errors.New("can not query random server port")
			}

			location := mullvadObject.Locations[relay.Location]
			server := WireguardServer{
				Pubkey:  relay.Pubkey,
				IPv4:    relay.IPv4Addr,
				Port:    *port,
				City:    location.City,
				Country: location.Country,
			}

			all_servers = append(all_servers, server)
		}
	}

	return all_servers, nil
}

// Add new public key to the account. This will create new mullvad device
func (m *WireguardProvider) AddPubkey(key string) error {
	url := m.URL("accounts", "v1", "devices")
	request := mullvadDeviceRequest{
		Pubkey:    key,
		HijackDNS: HijackDNSOption,
	}
	device := mullvadDevice{}

	err := m.APIRequest("POST", url, true, request, &device)

	if err != nil {
		return fmt.Errorf("[AddPubkey] failed to add pubkey: %s", err)
	}

	return nil
}

// Remove existing public key from the account.
// This will delete mullvad device.
func (m *WireguardProvider) RemovePubkey(key string) error {
	// List all devices
	devices := []mullvadDevice{}
	url := m.URL("accounts", "v1", "devices")
	err := m.APIRequest("GET", url, true, nil, &devices)

	if err != nil {
		return fmt.Errorf("[RemovePubkey] failed to list devices: %s", err)
	}

	// Iterate all devices
	for _, device := range devices {
		if key == device.Pubkey {
			url := m.URL("accounts", "v1", "devices", device.Id)
			err := m.APIRequest("DELETE", url, true, nil, nil)

			if err != nil {
				return fmt.Errorf("[RemovePubkey] failed to delete device: %s", err)
			}
		}
	}

	return nil
}

// Iterate all devices and fetch an address for a matching pubkey
func (m *WireguardProvider) GetAddress(key string) (string, error) {
	// List all devices
	devices := []mullvadDevice{}
	url := m.URL("accounts", "v1", "devices")
	err := m.APIRequest("GET", url, true, nil, &devices)

	if err != nil {
		return "", fmt.Errorf("[GetAddress] failed to list devices: %s", err)
	}

	// Iterate all devices
	for _, device := range devices {
		if key == device.Pubkey {
			return device.IPv4Address, nil
		}
	}

	// Nothing has been found
	return "", errors.New("[GetAddress] no matching device found")
}
