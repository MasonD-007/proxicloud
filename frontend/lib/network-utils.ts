/**
 * Network validation utilities for ProxiCloud frontend
 */

/**
 * Validates if a string is a valid CIDR notation (e.g., "10.0.1.0/24")
 * Checks that:
 * - The format is valid IP/mask
 * - The IP is the network address (not a host address)
 */
export function validateCIDR(cidr: string): { valid: boolean; error?: string } {
  if (!cidr || cidr.trim() === '') {
    return { valid: false, error: 'Subnet cannot be empty' };
  }

  const parts = cidr.split('/');
  if (parts.length !== 2) {
    return { valid: false, error: 'Invalid CIDR format. Use format: 10.0.1.0/24' };
  }

  const [ip, maskStr] = parts;
  const mask = parseInt(maskStr, 10);

  // Validate mask
  if (isNaN(mask) || mask < 0 || mask > 32) {
    return { valid: false, error: 'Invalid subnet mask. Must be between 0 and 32' };
  }

  // Validate IP format
  const ipParts = ip.split('.');
  if (ipParts.length !== 4) {
    return { valid: false, error: 'Invalid IP address format' };
  }

  const octets: number[] = [];
  for (const part of ipParts) {
    const octet = parseInt(part, 10);
    if (isNaN(octet) || octet < 0 || octet > 255) {
      return { valid: false, error: 'Invalid IP address. Each octet must be 0-255' };
    }
    octets.push(octet);
  }

  // Calculate network address
  const ipNum = (octets[0] << 24) | (octets[1] << 16) | (octets[2] << 8) | octets[3];
  const maskNum = mask === 0 ? 0 : ~((1 << (32 - mask)) - 1);
  const networkNum = ipNum & maskNum;

  // Check if the provided IP is the network address
  if (ipNum !== networkNum) {
    const networkOctets = [
      (networkNum >>> 24) & 0xff,
      (networkNum >>> 16) & 0xff,
      (networkNum >>> 8) & 0xff,
      networkNum & 0xff,
    ];
    const networkAddr = networkOctets.join('.');
    return {
      valid: false,
      error: `Not a network address. Use ${networkAddr}/${mask} instead`,
    };
  }

  return { valid: true };
}

/**
 * Validates if a gateway IP is valid and within the given subnet
 */
export function validateGatewayInSubnet(
  subnet: string,
  gateway: string
): { valid: boolean; error?: string } {
  if (!gateway || gateway.trim() === '') {
    return { valid: false, error: 'Gateway cannot be empty' };
  }

  // First validate subnet
  const subnetValidation = validateCIDR(subnet);
  if (!subnetValidation.valid) {
    return { valid: false, error: `Invalid subnet: ${subnetValidation.error}` };
  }

  // Validate gateway IP format
  const gatewayParts = gateway.split('.');
  if (gatewayParts.length !== 4) {
    return { valid: false, error: 'Invalid gateway IP format' };
  }

  const gatewayOctets: number[] = [];
  for (const part of gatewayParts) {
    const octet = parseInt(part, 10);
    if (isNaN(octet) || octet < 0 || octet > 255) {
      return { valid: false, error: 'Invalid gateway IP. Each octet must be 0-255' };
    }
    gatewayOctets.push(octet);
  }

  // Parse subnet
  const [subnetIP, maskStr] = subnet.split('/');
  const mask = parseInt(maskStr, 10);
  const subnetParts = subnetIP.split('.');
  const subnetOctets = subnetParts.map((p) => parseInt(p, 10));

  // Calculate network and broadcast addresses
  const subnetNum = (subnetOctets[0] << 24) | (subnetOctets[1] << 16) | (subnetOctets[2] << 8) | subnetOctets[3];
  const maskNum = ~((1 << (32 - mask)) - 1);
  const networkNum = subnetNum & maskNum;
  const broadcastNum = networkNum | ~maskNum;

  const gatewayNum = (gatewayOctets[0] << 24) | (gatewayOctets[1] << 16) | (gatewayOctets[2] << 8) | gatewayOctets[3];

  // Check if gateway is within subnet range
  if (gatewayNum < networkNum || gatewayNum > broadcastNum) {
    return { valid: false, error: 'Gateway is not within the subnet range' };
  }

  // Check if gateway is the network address
  if (gatewayNum === networkNum) {
    return { valid: false, error: 'Gateway cannot be the network address' };
  }

  // Check if gateway is the broadcast address
  if (gatewayNum === broadcastNum) {
    return { valid: false, error: 'Gateway cannot be the broadcast address' };
  }

  return { valid: true };
}

/**
 * Calculates and displays a preview of the DHCP range for a given subnet and gateway
 * Uses the entire available IP range except network address, broadcast, and gateway
 */
export function calculateDHCPRangePreview(subnet: string, gateway: string): string {
  // Validate inputs first
  const subnetValidation = validateCIDR(subnet);
  if (!subnetValidation.valid) {
    return '';
  }

  const gatewayValidation = validateGatewayInSubnet(subnet, gateway);
  if (!gatewayValidation.valid) {
    return '';
  }

  // Parse subnet
  const [subnetIP, maskStr] = subnet.split('/');
  const mask = parseInt(maskStr, 10);
  const subnetParts = subnetIP.split('.');
  const subnetOctets = subnetParts.map((p) => parseInt(p, 10));

  // Parse gateway
  const gatewayParts = gateway.split('.');
  const gatewayOctets = gatewayParts.map((p) => parseInt(p, 10));

  // Calculate network, broadcast, and gateway as numbers
  const networkNum = (subnetOctets[0] << 24) | (subnetOctets[1] << 16) | (subnetOctets[2] << 8) | subnetOctets[3];
  const gatewayNum = (gatewayOctets[0] << 24) | (gatewayOctets[1] << 16) | (gatewayOctets[2] << 8) | gatewayOctets[3];
  const totalHosts = 1 << (32 - mask);
  const broadcastNum = networkNum + totalHosts - 1;

  // Start from first usable IP (network + 1)
  let startIPNum = networkNum + 1;

  // End at last usable IP (broadcast - 1)
  let endIPNum = broadcastNum - 1;

  // Skip gateway if it's at the start
  if (startIPNum === gatewayNum) {
    startIPNum++;
  }

  // Skip gateway if it's at the end
  if (endIPNum === gatewayNum) {
    endIPNum--;
  }

  // Calculate number of IPs in range
  const dhcpSize = endIPNum - startIPNum + 1;

  const startIP = [
    (startIPNum >>> 24) & 0xff,
    (startIPNum >>> 16) & 0xff,
    (startIPNum >>> 8) & 0xff,
    startIPNum & 0xff,
  ].join('.');

  const endIP = [
    (endIPNum >>> 24) & 0xff,
    (endIPNum >>> 16) & 0xff,
    (endIPNum >>> 8) & 0xff,
    endIPNum & 0xff,
  ].join('.');

  return `${startIP} - ${endIP} (${dhcpSize} IPs)`;
}

/**
 * Provides helpful examples for subnet configuration
 */
export function getSubnetExamples(): string[] {
  return [
    '10.0.1.0/24 - 254 usable IPs',
    '192.168.0.0/24 - 254 usable IPs',
    '172.16.0.0/16 - 65,534 usable IPs',
  ];
}
