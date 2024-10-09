// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"log"
	"net"
)

// DefaultHttpBindAddress defines the default IP address for the HTTP server.
const DefaultHttpBindAddress = "0.0.0.0"

// ValidateHTTPAddress validates if the provided HTTP address is valid and
// assigned to an interface.
func ValidateHTTPAddress(httpAddress string) error {
	if httpAddress == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if httpAddress == DefaultHttpBindAddress {
		return fmt.Errorf("default bind address %s is not allowed", DefaultHttpBindAddress)
	}
	if net.ParseIP(httpAddress) == nil {
		return fmt.Errorf("invalid IP address format: %s", httpAddress)
	}
	if !IsIPInInterfaces(httpAddress) {
		log.Printf("[WARN] %s is not assigned to an interface", httpAddress)
	}
	return nil
}

// IsIPInInterfaces checks if the provided IP address is assigned to any
// interface in the system.
func IsIPInInterfaces(ipStr string) bool {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			parsedIP := net.ParseIP(ipStr)
			if ip.Equal(parsedIP) {
				return true
			}
		}
	}

	return false
}
