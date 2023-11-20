package providers

import (
	"errors"
	"math/rand"
)

// How many times to try to guess new upstream server
const maxConnectionGuesses = 33

type randomItem interface {
	WireguardServer | string | int
}

// Get a random element from the list
func GetRandomElement[T randomItem](elements []T) *T {
	total := len(elements)

	if total == 0 {
		return nil
	}

	if total == 1 {
		return &elements[1]
	}

	index := rand.Intn(total)

	return &elements[index]
}

// Given a list of servers, select one of them randomly for a particular location
func GuessNewUpstream(Servers *ServersState, Previous *WireguardServer, Location string) (WireguardServer, error) {
	if Servers == nil {
		return WireguardServer{}, errors.New("servers state is nil")
	}

	if len(Servers.Available) == 0 || len(Servers.Locations) == 0 {
		return WireguardServer{}, errors.New("no upstream servers available")
	}

	var available []WireguardServer
	var all_servers []WireguardServer

	// Filter available servers by location
	for location, servers := range Servers.Available {
		if location == Location {
			available = append(available, servers...)
		}
		all_servers = append(all_servers, servers...)
	}

	// If there was no matches, use any servers
	if len(available) == 0 {
		available = all_servers
	}

	// If there's only 1 server available, just return it
	if len(available) == 1 {
		return available[0], nil
	}

	// If there was no previous server (first run), compare against empty one
	if Previous == nil {
		Previous = &WireguardServer{}
	}

	// So apparently there are some servers available,
	// shuffle them until new one is selected (up to maxConnectionGuesses times)
	totalGuesses := 0

	for {
		if totalGuesses > maxConnectionGuesses {
			return WireguardServer{}, errors.New("could not guess new upstream after many tries")
		}

		candidate := GetRandomElement(available)

		// Force new city
		if candidate.City != Previous.City {
			return *candidate, nil
		}

		totalGuesses++
	}
}
