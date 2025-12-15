package proxmox

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateSDNIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		projectID   string
		projectName string
		wantPrefix  string
		wantLen     int
	}{
		{
			name:        "Normal project ID",
			projectID:   "a1b2c3d4e5f6",
			projectName: "Test Project",
			wantPrefix:  "prja1b2c",
			wantLen:     8,
		},
		{
			name:        "Short project ID",
			projectID:   "abc",
			projectName: "Test Project",
			wantPrefix:  "prj",
			wantLen:     8,
		},
		{
			name:        "Minimum valid project ID",
			projectID:   "abcde",
			projectName: "Test Project",
			wantPrefix:  "prjabcde",
			wantLen:     8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSDNIdentifier(tt.projectID, tt.projectName)

			if !strings.HasPrefix(got, "prj") {
				t.Errorf("GenerateSDNIdentifier() = %v, want prefix 'prj'", got)
			}

			if len(got) != tt.wantLen {
				t.Errorf("GenerateSDNIdentifier() length = %v, want %v", len(got), tt.wantLen)
			}

			if tt.projectID != "" && len(tt.projectID) >= 5 {
				if got != tt.wantPrefix {
					t.Errorf("GenerateSDNIdentifier() = %v, want %v", got, tt.wantPrefix)
				}
			}
		})
	}
}

func TestIsValidCIDR(t *testing.T) {
	tests := []struct {
		name string
		cidr string
		want bool
	}{
		{
			name: "Valid /24 network",
			cidr: "10.0.1.0/24",
			want: true,
		},
		{
			name: "Valid /16 network",
			cidr: "192.168.0.0/16",
			want: true,
		},
		{
			name: "Valid /8 network",
			cidr: "10.0.0.0/8",
			want: true,
		},
		{
			name: "Invalid - host address in /24",
			cidr: "10.0.1.5/24",
			want: false,
		},
		{
			name: "Invalid - host address in /16",
			cidr: "192.168.1.0/16",
			want: false,
		},
		{
			name: "Invalid - no mask",
			cidr: "10.0.1.0",
			want: false,
		},
		{
			name: "Invalid - empty string",
			cidr: "",
			want: false,
		},
		{
			name: "Invalid - malformed",
			cidr: "not-an-ip/24",
			want: false,
		},
		{
			name: "Valid /32 host",
			cidr: "10.0.1.1/32",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCIDR(tt.cidr); got != tt.want {
				t.Errorf("IsValidCIDR(%q) = %v, want %v", tt.cidr, got, tt.want)
			}
		})
	}
}

func TestValidateGatewayInSubnet(t *testing.T) {
	tests := []struct {
		name    string
		subnet  string
		gateway string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Valid gateway in /24",
			subnet:  "10.0.1.0/24",
			gateway: "10.0.1.1",
			wantErr: false,
		},
		{
			name:    "Valid gateway mid-range",
			subnet:  "192.168.0.0/16",
			gateway: "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "Invalid - gateway outside subnet",
			subnet:  "10.0.1.0/24",
			gateway: "10.0.2.1",
			wantErr: true,
			errMsg:  "not within subnet",
		},
		{
			name:    "Invalid - gateway is network address",
			subnet:  "10.0.1.0/24",
			gateway: "10.0.1.0",
			wantErr: true,
			errMsg:  "network address",
		},
		{
			name:    "Invalid - gateway is broadcast",
			subnet:  "10.0.1.0/24",
			gateway: "10.0.1.255",
			wantErr: true,
			errMsg:  "broadcast address",
		},
		{
			name:    "Invalid - empty subnet",
			subnet:  "",
			gateway: "10.0.1.1",
			wantErr: true,
			errMsg:  "subnet cannot be empty",
		},
		{
			name:    "Invalid - empty gateway",
			subnet:  "10.0.1.0/24",
			gateway: "",
			wantErr: true,
			errMsg:  "gateway cannot be empty",
		},
		{
			name:    "Invalid - malformed subnet",
			subnet:  "not-a-cidr",
			gateway: "10.0.1.1",
			wantErr: true,
			errMsg:  "invalid subnet CIDR",
		},
		{
			name:    "Invalid - malformed gateway",
			subnet:  "10.0.1.0/24",
			gateway: "not-an-ip",
			wantErr: true,
			errMsg:  "invalid gateway IP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGatewayInSubnet(tt.subnet, tt.gateway)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateGatewayInSubnet() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateGatewayInSubnet() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateGatewayInSubnet() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestCalculateDHCPRange(t *testing.T) {
	tests := []struct {
		name       string
		subnet     string
		gateway    string
		wantErr    bool
		wantPrefix string
	}{
		{
			name:       "Valid /24 network",
			subnet:     "10.0.1.0/24",
			gateway:    "10.0.1.1",
			wantErr:    false,
			wantPrefix: "start-address=",
		},
		{
			name:       "Valid /16 network",
			subnet:     "192.168.0.0/16",
			gateway:    "192.168.0.1",
			wantErr:    false,
			wantPrefix: "start-address=",
		},
		{
			name:    "Invalid - malformed subnet",
			subnet:  "not-a-cidr",
			gateway: "10.0.1.1",
			wantErr: true,
		},
		{
			name:    "Invalid - malformed gateway",
			subnet:  "10.0.1.0/24",
			gateway: "not-an-ip",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateDHCPRange(tt.subnet, tt.gateway)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CalculateDHCPRange() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("CalculateDHCPRange() unexpected error = %v", err)
				return
			}

			if !strings.HasPrefix(got, tt.wantPrefix) {
				t.Errorf("CalculateDHCPRange() = %v, want prefix %v", got, tt.wantPrefix)
			}

			// Verify the format contains both start and end
			if !strings.Contains(got, "start-address=") || !strings.Contains(got, "end-address=") {
				t.Errorf("CalculateDHCPRange() = %v, want format with start-address and end-address", got)
			}

			// Parse and verify the range is valid
			startIP, endIP, err := ParseDHCPRange(got)
			if err != nil {
				t.Errorf("CalculateDHCPRange() produced invalid range: %v", err)
			}

			if startIP == "" || endIP == "" {
				t.Errorf("CalculateDHCPRange() produced empty start or end IP")
			}
		})
	}
}

func TestParseDHCPRange(t *testing.T) {
	tests := []struct {
		name      string
		dhcpRange string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			name:      "Valid range",
			dhcpRange: "start-address=10.0.1.100,end-address=10.0.1.200",
			wantStart: "10.0.1.100",
			wantEnd:   "10.0.1.200",
			wantErr:   false,
		},
		{
			name:      "Valid range with spaces",
			dhcpRange: "start-address=10.0.1.100, end-address=10.0.1.200",
			wantStart: "10.0.1.100",
			wantEnd:   "10.0.1.200",
			wantErr:   false,
		},
		{
			name:      "Invalid - empty string",
			dhcpRange: "",
			wantErr:   true,
		},
		{
			name:      "Invalid - missing end",
			dhcpRange: "start-address=10.0.1.100",
			wantErr:   true,
		},
		{
			name:      "Invalid - missing start",
			dhcpRange: "end-address=10.0.1.200",
			wantErr:   true,
		},
		{
			name:      "Invalid - malformed",
			dhcpRange: "not-a-valid-range",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := ParseDHCPRange(tt.dhcpRange)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDHCPRange() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDHCPRange() unexpected error = %v", err)
				return
			}

			if gotStart != tt.wantStart {
				t.Errorf("ParseDHCPRange() start = %v, want %v", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("ParseDHCPRange() end = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestIPConversions(t *testing.T) {
	tests := []struct {
		name   string
		ipStr  string
		wantIP string
	}{
		{
			name:   "Simple IP",
			ipStr:  "10.0.1.1",
			wantIP: "10.0.1.1",
		},
		{
			name:   "Network address",
			ipStr:  "192.168.0.0",
			wantIP: "192.168.0.0",
		},
		{
			name:   "Broadcast",
			ipStr:  "10.0.1.255",
			wantIP: "10.0.1.255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert string to IP to uint32
			ip := parseIPv4(tt.ipStr)
			if ip == nil {
				t.Fatalf("Failed to parse IP %s", tt.ipStr)
			}

			uint32Val := ipToUint32(ip)

			// Convert back to IP
			gotIP := uint32ToIP(uint32Val)

			if gotIP.String() != tt.wantIP {
				t.Errorf("IP conversion round-trip: got %v, want %v", gotIP.String(), tt.wantIP)
			}
		})
	}
}

// Helper function to parse IPv4 addresses for testing
func parseIPv4(s string) []byte {
	var parts [4]byte
	_, err := fmt.Sscanf(s, "%d.%d.%d.%d", &parts[0], &parts[1], &parts[2], &parts[3])
	if err != nil {
		return nil
	}
	return parts[:]
}
