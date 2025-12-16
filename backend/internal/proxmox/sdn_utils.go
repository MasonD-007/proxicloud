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
// Returns a string in the format "start-address=10.0.0.2,end-address=10.0.0.254"
// The range uses all available IPs in the subnet except network address, broadcast, and gateway
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

	// Get the subnet size
	ones, bits := ipNet.Mask.Size()
	if bits != 32 {
		// Only support IPv4 for now
		return "", fmt.Errorf("only IPv4 subnets are supported")
	}

	// Calculate total hosts
	totalHosts := 1 << uint(bits-ones)

	// Convert network address and gateway to uint32 for easier math
	networkIP := ipToUint32(ipNet.IP)
	gatewayIPUint32 := ipToUint32(gatewayIP.To4())

	// Calculate broadcast address
	broadcastIP := networkIP + uint32(totalHosts) - 1

	// Start from first usable IP (network + 1)
	startIPUint32 := networkIP + 1

	// End at last usable IP (broadcast - 1)
	endIPUint32 := broadcastIP - 1

	// Skip gateway IP if it's at the start
	if startIPUint32 == gatewayIPUint32 {
		startIPUint32++
	}

	// Skip gateway IP if it's at the end
	if endIPUint32 == gatewayIPUint32 {
		endIPUint32--
	}

	// Convert back to IPs
	startIP := uint32ToIP(startIPUint32)
	endIP := uint32ToIP(endIPUint32)

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
