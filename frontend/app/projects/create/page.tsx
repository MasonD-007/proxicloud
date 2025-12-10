'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Network } from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import { createProject } from '@/lib/api';
import type { CreateProjectRequest } from '@/lib/types';

export default function CreateProjectPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    description: '',
    tags: [],
    network: undefined,
  });
  const [tagInput, setTagInput] = useState('');
  const [enableNetwork, setEnableNetwork] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      setError('Project name is required');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      await createProject(formData);
      router.push('/projects');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project');
    } finally {
      setLoading(false);
    }
  }

  function addTag() {
    const tag = tagInput.trim();
    if (tag && !formData.tags?.includes(tag)) {
      setFormData({
        ...formData,
        tags: [...(formData.tags || []), tag],
      });
      setTagInput('');
    }
  }

  function removeTag(tag: string) {
    setFormData({
      ...formData,
      tags: formData.tags?.filter(t => t !== tag) || [],
    });
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="flex items-center gap-4">
        <Link href="/projects">
          <Button variant="outline" size="sm">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>
        <h1 className="text-3xl font-bold text-text-primary">Create Project</h1>
      </div>

      <Card>
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Project Name *
            </label>
            <Input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="my-awesome-project"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Project description..."
              rows={4}
              className="w-full px-4 py-2 bg-background-elevated border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary text-text-primary"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-primary mb-2">
              Tags
            </label>
            <div className="flex gap-2 mb-2">
              <Input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyPress={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    addTag();
                  }
                }}
                placeholder="Add tag..."
              />
              <Button type="button" onClick={addTag} variant="outline">
                Add
              </Button>
            </div>
            {formData.tags && formData.tags.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {formData.tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center gap-1 px-3 py-1 bg-background-elevated border border-border rounded-full text-sm text-text-primary"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => removeTag(tag)}
                      className="text-text-muted hover:text-error"
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Network Configuration Section */}
          <div className="border-t border-border pt-6">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center gap-2">
                <Network className="w-5 h-5 text-primary" />
                <label className="text-sm font-medium text-text-primary">
                  Network Configuration (Optional)
                </label>
              </div>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={enableNetwork}
                  onChange={(e) => {
                    setEnableNetwork(e.target.checked);
                    if (!e.target.checked) {
                      setFormData({ ...formData, network: undefined });
                    } else {
                      setFormData({ 
                        ...formData, 
                        network: { subnet: '', gateway: '', nameserver: '' } 
                      });
                    }
                  }}
                  className="w-4 h-4"
                />
                <span className="text-sm text-text-secondary">Enable</span>
              </label>
            </div>

            {enableNetwork && (
              <div className="space-y-4 pl-7">
                <p className="text-sm text-text-muted mb-4">
                  Configure a default network for this project. New containers will automatically use these settings.
                </p>
                
                <div>
                  <label className="block text-sm font-medium text-text-primary mb-2">
                    Subnet (CIDR)
                  </label>
                  <Input
                    type="text"
                    value={formData.network?.subnet || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      network: { ...formData.network, subnet: e.target.value } 
                    })}
                    placeholder="e.g., 192.168.1.0/24"
                  />
                  <p className="text-xs text-text-muted mt-1">
                    The network subnet in CIDR notation
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-text-primary mb-2">
                    Gateway
                  </label>
                  <Input
                    type="text"
                    value={formData.network?.gateway || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      network: { ...formData.network, gateway: e.target.value } 
                    })}
                    placeholder="e.g., 192.168.1.1"
                  />
                  <p className="text-xs text-text-muted mt-1">
                    Default gateway for containers in this project
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-text-primary mb-2">
                    DNS Nameserver
                  </label>
                  <Input
                    type="text"
                    value={formData.network?.nameserver || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      network: { ...formData.network, nameserver: e.target.value } 
                    })}
                    placeholder="e.g., 8.8.8.8"
                  />
                  <p className="text-xs text-text-muted mt-1">
                    DNS server for name resolution
                  </p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-text-primary mb-2">
                    VLAN Tag (Optional)
                  </label>
                  <Input
                    type="number"
                    min="1"
                    max="4094"
                    value={formData.network?.vlan_tag || ''}
                    onChange={(e) => setFormData({ 
                      ...formData, 
                      network: { 
                        ...formData.network, 
                        vlan_tag: e.target.value ? parseInt(e.target.value) : undefined 
                      } 
                    })}
                    placeholder="e.g., 100"
                  />
                  <p className="text-xs text-text-muted mt-1">
                    VLAN tag for network isolation (1-4094)
                  </p>
                </div>
              </div>
            )}
          </div>

          {error && (
            <div className="p-4 bg-error/10 border border-error rounded-lg text-error text-sm">
              {error}
            </div>
          )}

          <div className="flex gap-4">
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating...' : 'Create Project'}
            </Button>
            <Link href="/projects">
              <Button type="button" variant="outline">
                Cancel
              </Button>
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
