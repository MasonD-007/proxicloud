package proxmox

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// GenerateSDNIdentifier generates a unique 8-character identifier for SDN zones/vnets
// Format: "prj" + first 5 chars of projectID hex
// Example: projectID="a1b2c3d4..." → "prja1b2c"
func GenerateSDNIdentifier(projectID string, projectName string) string {
	// Use project ID (already unique) to generate zone/vnet ID
	if len(projectID) >= 5 {
		return "prj" + projectID[:5]
	}

	// Fallback: hash the name if projectID is too short
	hash := sha256.Sum256([]byte(projectName))
	return "prj" + hex.EncodeToString(hash[:])[:5]
}

// IsValidCIDR validates that a string is a valid CIDR notation
func IsValidCIDR(cidr string) bool {
	if cidr == "" {
		return false
	}

	// Parse the CIDR to get the IP and mask
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	// Verify that the provided IP is the network address (not a host address)
	// net.ParseCIDR returns the original IP in 'ip' and the network address in 'ipNet.IP'
	// For example, "10.0.1.0/24" is valid, but "10.0.1.5/24" is not
	// We need to check if the original IP matches the network address
	if !ip.Equal(ipNet.IP) {
		return false
	}

	return true
}

// ValidateGatewayInSubnet validates that a gateway IP is within the given subnet
func ValidateGatewayInSubnet(subnet string, gateway string) error {
	if subnet == "" {
		return fmt.Errorf("subnet cannot be empty")
	}
	if gateway == "" {
		return fmt.Errorf("gateway cannot be empty")
	}

	// Parse subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet CIDR: %w", err)
	}

	// Parse gateway IP
	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return fmt.Errorf("invalid gateway IP address")
	}

	// Check if gateway is within subnet
	if !ipNet.Contains(gatewayIP) {
		return fmt.Errorf("gateway %s is not within subnet %s", gateway, subnet)
	}

	// Check if gateway is the network address
	if gatewayIP.Equal(ipNet.IP) {
		return fmt.Errorf("gateway cannot be the network address (%s)", ipNet.IP.String())
	}

	// Check if gateway is the broadcast address (for IPv4)
	if gatewayIP.To4() != nil {
		broadcast := make(net.IP, len(ipNet.IP))
		copy(broadcast, ipNet.IP)
		for i := range broadcast {
			broadcast[i] |= ^ipNet.Mask[i]
		}
		if gatewayIP.Equal(broadcast) {
			return fmt.Errorf("gateway cannot be the broadcast address (%s)", broadcast.String())
		}
	}

	return nil
}

// CalculateDHCPRange calculates a DHCP range for Proxmox SDN
// Returns a string in the format "start-ip=10.0.1.100,end-ip=10.0.1.200"
// The range excludes the gateway and uses the upper portion of the subnet
func CalculateDHCPRange(subnet string, gateway string) (string, error) {
	// Parse subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("invalid subnet CIDR: %w", err)
	}

	// Parse gateway IP
	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return "", fmt.Errorf("invalid gateway IP address")
	}

	// Calculate the usable IP range
	// Network address is ipNet.IP
	// Broadcast address (for IPv4) is calculated below

	// For simplicity, we'll allocate the upper 50% of the subnet for DHCP
	// For example, in a /24 network (254 usable IPs):
	// - Gateway typically at .1
	// - DHCP range from .100 to .200 (100 IPs)

	// Get the subnet size
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		// Only support IPv4 for now
		return "", fmt.Errorf("only IPv4 subnets are supported")
	}

	// Calculate total hosts
	totalHosts := 1 << uint(bits-ones)

	// Calculate usable hosts (subtract network and broadcast)
	usableHosts := totalHosts - 2

	// Calculate DHCP range (upper 40% of usable IPs, or at least 10 IPs)
	dhcpSize := usableHosts * 40 / 100
	if dhcpSize < 10 && usableHosts >= 10 {
		dhcpSize = 10
	}
	if dhcpSize > usableHosts {
		dhcpSize = usableHosts
	}

	// Start DHCP range from a reasonable offset (e.g., .100 or 60% into the range)
	startOffset := usableHosts * 60 / 100
	if startOffset < 10 {
		startOffset = 10
	}

	// Ensure we don't exceed the subnet
	if startOffset+dhcpSize > usableHosts {
		startOffset = usableHosts - dhcpSize
		if startOffset < 1 {
			startOffset = 1
		}
	}

	// Convert network address to uint32 for easier math
	networkIP := ipToUint32(ipNet.IP)

	// Calculate start and end IPs
	startIP := uint32ToIP(networkIP + uint32(startOffset))
	endIP := uint32ToIP(networkIP + uint32(startOffset+dhcpSize-1))

	// Format the range string for Proxmox SDN
	return fmt.Sprintf("start-address=%s,end-address=%s", startIP.String(), endIP.String()), nil
}

// ipToUint32 converts an IPv4 address to uint32
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// uint32ToIP converts a uint32 to an IPv4 address
func uint32ToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}

// ParseDHCPRange parses a DHCP range string and returns start and end IPs
// Example input: "start-address=10.0.1.100,end-address=10.0.1.200"
func ParseDHCPRange(dhcpRange string) (startIP, endIP string, err error) {
	if dhcpRange == "" {
		return "", "", fmt.Errorf("DHCP range is empty")
	}

	parts := strings.Split(dhcpRange, ",")
	for _, part := range parts {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "start-address":
			startIP = value
		case "end-address":
			endIP = value
		}
	}

	if startIP == "" || endIP == "" {
		return "", "", fmt.Errorf("invalid DHCP range format: %s", dhcpRange)
	}

	return startIP, endIP, nil
}
