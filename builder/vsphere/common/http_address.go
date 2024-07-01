// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"net"
)

// ValidateHTTPAddress validates if the provided HTTP address is valid and
// assigned to any interface.
func ValidateHTTPAddress(httpAddress string) error {
	if httpAddress != "" && httpAddress != "0.0.0.0" {
		if !IsIPInInterfaces(httpAddress) {
			return fmt.Errorf("'http_bind_address' %s is not assigned to any interface", httpAddress)
		}
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
