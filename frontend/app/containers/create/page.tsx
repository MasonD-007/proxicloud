'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft, Server } from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import { createContainer, getTemplates, getProjects, getProject } from '@/lib/api';
import type { CreateContainerRequest, Template, Project } from '@/lib/types';

export default function CreateContainerPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const projectIdParam = searchParams.get('project_id');
  
  const [loading, setLoading] = useState(false);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [projectsLoading, setProjectsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<CreateContainerRequest>({
    vmid: undefined, // Optional VMID
    hostname: '',
    cores: 2,
    memory: 1024,
    disk: 8,
    ostemplate: '',
    password: '',
    ssh_keys: '',
    start_on_boot: false,
    unprivileged: true,
    project_id: undefined,
    ip_address: '',
    gateway: '',
    nameserver: '8.8.8.8',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [useDHCP, setUseDHCP] = useState(true);

  useEffect(() => {
    loadTemplates();
    loadProjects();
  }, []);

  async function loadTemplates() {
    try {
      setTemplatesLoading(true);
      const data = await getTemplates();
      setTemplates(data);
      if (data.length > 0) {
        setFormData((prev) => ({ ...prev, ostemplate: data[0].volid }));
      }
    } catch (err) {
      console.error('Failed to load templates:', err);
    } finally {
      setTemplatesLoading(false);
    }
  }

  async function loadProjects() {
    try {
      setProjectsLoading(true);
      const data = await getProjects();
      setProjects(data);
    } catch (err) {
      console.error('Failed to load projects:', err);
    } finally {
      setProjectsLoading(false);
    }
  }

  const loadProjectDetails = useCallback(async (projectId: string) => {
    try {
      const project = await getProject(projectId);
      setSelectedProject(project);
      
      // Set project_id in form
      setFormData((prev) => ({ ...prev, project_id: projectId }));
      
      // Auto-populate network settings if project has network config
      if (project.network) {
        setUseDHCP(false);
        const network = project.network;
        
        // Calculate next available IP from subnet if provided
        if (network.subnet) {
          // Extract base IP from subnet (e.g., "192.168.1.0/24" -> "192.168.1.")
          const subnetParts = network.subnet.split('/')[0].split('.');
          if (subnetParts.length === 4) {
            // Suggest the next IP (e.g., 192.168.1.10/24)
            const suggestedIP = `${subnetParts[0]}.${subnetParts[1]}.${subnetParts[2]}.10/${network.subnet.split('/')[1]}`;
            setFormData((prev) => ({ ...prev, ip_address: suggestedIP }));
          }
        }
        
        if (network.gateway) {
          setFormData((prev) => ({ ...prev, gateway: network.gateway }));
        }
        
        if (network.nameserver) {
          setFormData((prev) => ({ ...prev, nameserver: network.nameserver }));
        }
      }
    } catch (err) {
      console.error('Failed to load project details:', err);
    }
  }, []);

  // Load project details if project_id is in query params
  useEffect(() => {
    if (projectIdParam && projects.length > 0) {
      loadProjectDetails(projectIdParam);
    }
  }, [projectIdParam, projects, loadProjectDetails]);


  function validateForm(): boolean {
    const newErrors: Record<string, string> = {};

    if (!formData.hostname || formData.hostname.length < 2) {
      newErrors.hostname = 'Hostname must be at least 2 characters';
    }
    if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(formData.hostname)) {
      newErrors.hostname = 'Hostname must contain only lowercase letters, numbers, and hyphens';
    }
    if (!formData.ostemplate) {
      newErrors.ostemplate = 'Please select an OS template';
    }
    if (formData.vmid !== undefined && (formData.vmid < 100 || formData.vmid > 999999999)) {
      newErrors.vmid = 'VMID must be between 100 and 999999999';
    }
    if (formData.cores < 1 || formData.cores > 128) {
      newErrors.cores = 'CPU cores must be between 1 and 128';
    }
    if (formData.memory < 512 || formData.memory > 524288) {
      newErrors.memory = 'Memory must be between 512 MB and 512 GB';
    }
    if (formData.disk < 1 || formData.disk > 10240) {
      newErrors.disk = 'Disk size must be between 1 GB and 10 TB';
    }

    // Network validation (only if not using DHCP)
    if (!useDHCP) {
      if (!formData.ip_address) {
        newErrors.ip_address = 'IP address is required when not using DHCP';
      } else if (!/^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/.test(formData.ip_address)) {
        newErrors.ip_address = 'IP address must be in CIDR format (e.g., 192.168.1.100/24)';
      }
      
      if (formData.gateway && !/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.gateway)) {
        newErrors.gateway = 'Gateway must be a valid IP address (e.g., 192.168.1.1)';
      }
      
      if (formData.nameserver && !/^(\d{1,3}\.){3}\d{1,3}$/.test(formData.nameserver)) {
        newErrors.nameserver = 'Nameserver must be a valid IP address (e.g., 8.8.8.8)';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    try {
      setLoading(true);
      setError(null);
      
      // Prepare the request data
      const requestData: CreateContainerRequest = {
        ...formData,
        // Clear network fields if using DHCP
        ip_address: useDHCP ? undefined : formData.ip_address,
        gateway: useDHCP ? undefined : formData.gateway,
        nameserver: useDHCP ? undefined : formData.nameserver,
      };
      
      const result = await createContainer(requestData);
      router.push(`/containers/${result.vmid}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create container');
    } finally {
      setLoading(false);
    }
  }

  function handleInputChange(field: keyof CreateContainerRequest, value: string | number | boolean | undefined) {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/containers">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
        </Link>
        <div>
          <h1 className="text-3xl font-bold text-text-primary flex items-center gap-3">
            <Server className="w-8 h-8" />
            Create Container
          </h1>
          <p className="text-text-muted mt-1">
            Create a new LXC container on your Proxmox node
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {error && (
          <Card className="bg-error/10 border-error">
            <div className="text-error">{error}</div>
          </Card>
        )}

        <Card>
          <h2 className="text-xl font-semibold text-text-primary mb-4">Basic Configuration</h2>
          <div className="space-y-4">
            <Input
              label="Container ID (VMID)"
              placeholder="Leave empty for auto-assign"
              type="number"
              min="100"
              max="999999999"
              value={formData.vmid || ''}
              onChange={(e) => handleInputChange('vmid', e.target.value ? parseInt(e.target.value) : undefined)}
              error={errors.vmid}
            />
            <p className="text-sm text-text-muted -mt-2">
              Optional: Specify a custom container ID (100-999999999), or leave empty to auto-assign
            </p>

            <Input
              label="Hostname"
              placeholder="my-container"
              value={formData.hostname}
              onChange={(e) => handleInputChange('hostname', e.target.value)}
              error={errors.hostname}
              required
            />
            <p className="text-sm text-text-muted -mt-2">
              Use lowercase letters, numbers, and hyphens only
            </p>

            {templatesLoading ? (
              <div className="text-text-muted">Loading templates...</div>
            ) : templates.length > 0 ? (
              <Select
                label="OS Template"
                value={formData.ostemplate}
                onChange={(e) => handleInputChange('ostemplate', e.target.value)}
                error={errors.ostemplate}
                options={templates.map((t) => ({
                  value: t.volid,
                  label: t.volid.split('/').pop() || t.volid,
                }))}
                required
              />
            ) : (
              <div className="text-warning">
                No templates found. Please upload templates to your Proxmox storage.
              </div>
            )}

            {/* Project Selection */}
            {projectsLoading ? (
              <div className="text-text-muted">Loading projects...</div>
            ) : projects.length > 0 ? (
              <Select
                label="Project (Optional)"
                value={formData.project_id || ''}
                onChange={(e) => handleInputChange('project_id', e.target.value || undefined)}
                options={[
                  { value: '', label: 'No Project' },
                  ...projects.map((p) => ({
                    value: p.id,
                    label: p.name,
                  })),
                ]}
              />
            ) : (
              <div className="text-text-muted text-sm">
                No projects available. You can create projects to organize your containers.
              </div>
            )}
            {projects.length > 0 && (
              <p className="text-sm text-text-muted -mt-2">
                Optionally assign this container to a project for better organization
              </p>
            )}
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold text-text-primary mb-4">Resources</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <Input
                label="CPU Cores"
                type="number"
                min="1"
                max="128"
                value={formData.cores}
                onChange={(e) => handleInputChange('cores', parseInt(e.target.value))}
                error={errors.cores}
                required
              />
            </div>
            <div>
              <Input
                label="Memory (MB)"
                type="number"
                min="512"
                max="524288"
                step="256"
                value={formData.memory}
                onChange={(e) => handleInputChange('memory', parseInt(e.target.value))}
                error={errors.memory}
                required
              />
            </div>
            <div>
              <Input
                label="Disk Size (GB)"
                type="number"
                min="1"
                max="10240"
                value={formData.disk}
                onChange={(e) => handleInputChange('disk', parseInt(e.target.value))}
                error={errors.disk}
                required
              />
            </div>
          </div>
          <div className="mt-4 p-4 bg-surface-elevated rounded-lg">
            <div className="text-sm text-text-secondary space-y-1">
              <p><strong>Defaults:</strong> 2 cores, 1024 MB RAM, 8 GB disk</p>
              <p className="text-text-muted">Adjust resources based on your workload requirements</p>
            </div>
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold text-text-primary mb-4">Security</h2>
          <div className="space-y-4">
            <Input
              label="Root Password"
              type="password"
              placeholder="Leave empty for no password"
              value={formData.password}
              onChange={(e) => handleInputChange('password', e.target.value)}
            />
            <p className="text-sm text-text-muted -mt-2">
              Set a root password for the container (optional if using SSH keys)
            </p>

            <div>
              <label className="text-sm font-medium text-text-secondary block mb-2">
                SSH Public Keys
              </label>
              <textarea
                className="w-full bg-surface-elevated border border-border rounded-lg px-3 py-2 text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                rows={4}
                placeholder="ssh-rsa AAAAB3NzaC1yc2E... (one per line)"
                value={formData.ssh_keys}
                onChange={(e) => handleInputChange('ssh_keys', e.target.value)}
              />
              <p className="text-sm text-text-muted mt-1">
                Add SSH public keys for passwordless access (optional)
              </p>
            </div>
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold text-text-primary mb-4">Advanced Options</h2>
          <div className="space-y-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.start_on_boot}
                onChange={(e) => handleInputChange('start_on_boot', e.target.checked)}
                className="w-4 h-4 text-primary bg-surface-elevated border-border rounded focus:ring-2 focus:ring-primary"
              />
              <div>
                <div className="text-sm font-medium text-text-primary">Start on Boot</div>
                <div className="text-sm text-text-muted">
                  Automatically start this container when the Proxmox node boots
                </div>
              </div>
            </label>

            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={formData.unprivileged}
                onChange={(e) => handleInputChange('unprivileged', e.target.checked)}
                className="w-4 h-4 text-primary bg-surface-elevated border-border rounded focus:ring-2 focus:ring-primary"
              />
              <div>
                <div className="text-sm font-medium text-text-primary">Unprivileged Container</div>
                <div className="text-sm text-text-muted">
                  Use unprivileged container for better security (recommended)
                </div>
              </div>
            </label>
          </div>
        </Card>

        <Card>
          <h2 className="text-xl font-semibold text-text-primary mb-4">Network Configuration</h2>
          
          {selectedProject?.network && (
            <div className="mb-4 p-4 bg-primary/10 border border-primary/30 rounded-lg">
              <div className="flex items-start gap-2">
                <div className="text-primary font-medium text-sm">
                  Project Network Settings
                </div>
              </div>
              <div className="mt-2 text-sm text-text-secondary space-y-1">
                {selectedProject.network.subnet && (
                  <div>Subnet: <span className="font-mono text-text-primary">{selectedProject.network.subnet}</span></div>
                )}
                {selectedProject.network.gateway && (
                  <div>Gateway: <span className="font-mono text-text-primary">{selectedProject.network.gateway}</span></div>
                )}
                {selectedProject.network.nameserver && (
                  <div>DNS: <span className="font-mono text-text-primary">{selectedProject.network.nameserver}</span></div>
                )}
              </div>
              <p className="text-xs text-text-muted mt-2">
                Network settings have been auto-populated from the project configuration. You can modify them below if needed.
              </p>
            </div>
          )}
          
          <div className="space-y-4">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={useDHCP}
                onChange={(e) => {
                  setUseDHCP(e.target.checked);
                  if (e.target.checked) {
                    // Clear network fields when switching to DHCP
                    handleInputChange('ip_address', '');
                    handleInputChange('gateway', '');
                  }
                }}
                className="w-4 h-4 text-primary bg-surface-elevated border-border rounded focus:ring-2 focus:ring-primary"
              />
              <div>
                <div className="text-sm font-medium text-text-primary">Use DHCP</div>
                <div className="text-sm text-text-muted">
                  Automatically obtain IP address from DHCP server (recommended for most setups)
                </div>
              </div>
            </label>

            {!useDHCP && (
              <div className="space-y-4 pl-7 border-l-2 border-primary/30">
                <Input
                  label="IP Address (CIDR)"
                  placeholder="192.168.1.100/24"
                  value={formData.ip_address || ''}
                  onChange={(e) => handleInputChange('ip_address', e.target.value)}
                  error={errors.ip_address}
                  required={!useDHCP}
                />
                <p className="text-sm text-text-muted -mt-2">
                  IP address in CIDR notation (e.g., 192.168.1.100/24)
                </p>

                <Input
                  label="Gateway"
                  placeholder="192.168.1.1"
                  value={formData.gateway || ''}
                  onChange={(e) => handleInputChange('gateway', e.target.value)}
                  error={errors.gateway}
                />
                <p className="text-sm text-text-muted -mt-2">
                  Default gateway IP address (optional)
                </p>

                <Input
                  label="DNS Nameserver"
                  placeholder="8.8.8.8"
                  value={formData.nameserver || ''}
                  onChange={(e) => handleInputChange('nameserver', e.target.value)}
                  error={errors.nameserver}
                />
                <p className="text-sm text-text-muted -mt-2">
                  DNS server IP address (defaults to 8.8.8.8 if not specified)
                </p>
              </div>
            )}
          </div>
        </Card>

        <div className="flex items-center gap-4">
          <Button type="submit" disabled={loading || templatesLoading}>
            {loading ? 'Creating...' : 'Create Container'}
          </Button>
          <Link href="/containers">
            <Button variant="ghost" type="button">
              Cancel
            </Button>
          </Link>
        </div>
      </form>
    </div>
  );
}
