'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Server } from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import { createContainer, getTemplates } from '@/lib/api';
import type { CreateContainerRequest, Template } from '@/lib/types';

export default function CreateContainerPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [templates, setTemplates] = useState<Template[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<CreateContainerRequest>({
    hostname: '',
    cores: 2,
    memory: 1024,
    disk: 8,
    ostemplate: '',
    password: '',
    ssh_keys: '',
    start_on_boot: false,
    unprivileged: true,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    loadTemplates();
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
    if (formData.cores < 1 || formData.cores > 128) {
      newErrors.cores = 'CPU cores must be between 1 and 128';
    }
    if (formData.memory < 512 || formData.memory > 524288) {
      newErrors.memory = 'Memory must be between 512 MB and 512 GB';
    }
    if (formData.disk < 1 || formData.disk > 10240) {
      newErrors.disk = 'Disk size must be between 1 GB and 10 TB';
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
      const result = await createContainer(formData);
      router.push(`/containers/${result.vmid}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create container');
    } finally {
      setLoading(false);
    }
  }

  function handleInputChange(field: keyof CreateContainerRequest, value: string | number | boolean) {
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
