'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, HardDrive } from 'lucide-react';
import Link from 'next/link';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import { createVolume } from '@/lib/api';
import type { CreateVolumeRequest } from '@/lib/types';

export default function CreateVolumePage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<CreateVolumeRequest>({
    name: '',
    size: 10,
    storage: 'local-lvm',
    type: 'ssd',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  function validateForm(): boolean {
    const newErrors: Record<string, string> = {};

    if (!formData.name || formData.name.length < 2) {
      newErrors.name = 'Volume name must be at least 2 characters';
    }
    if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(formData.name)) {
      newErrors.name = 'Name must contain only lowercase letters, numbers, and hyphens';
    }
    if (formData.size < 1 || formData.size > 10240) {
      newErrors.size = 'Size must be between 1 GB and 10 TB';
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
      const volume = await createVolume(formData);
      router.push(`/volumes/${encodeURIComponent(volume.volid)}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create volume');
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center gap-4">
        <Link href="/volumes">
          <Button variant="ghost" size="sm">
            <ArrowLeft className="w-4 h-4" />
          </Button>
        </Link>
        <h1 className="text-3xl font-bold text-text-primary">Create Volume</h1>
      </div>

      <Card>
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="bg-error/10 border border-error text-error px-4 py-3 rounded-lg">
              {error}
            </div>
          )}

          {/* Volume Name */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-text-primary mb-2">
              Volume Name *
            </label>
            <Input
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value.toLowerCase() })}
              placeholder="my-volume"
              error={errors.name}
            />
            {errors.name && (
              <p className="mt-1 text-sm text-error">{errors.name}</p>
            )}
            <p className="mt-1 text-sm text-text-muted">
              Lowercase letters, numbers, and hyphens only
            </p>
          </div>

          {/* Size */}
          <div>
            <label htmlFor="size" className="block text-sm font-medium text-text-primary mb-2">
              Size (GB) *
            </label>
            <Input
              id="size"
              type="number"
              min="1"
              max="10240"
              value={formData.size}
              onChange={(e) => setFormData({ ...formData, size: parseInt(e.target.value) || 1 })}
              error={errors.size}
            />
            {errors.size && (
              <p className="mt-1 text-sm text-error">{errors.size}</p>
            )}
            <p className="mt-1 text-sm text-text-muted">
              Volume size in gigabytes (1 GB - 10 TB)
            </p>
          </div>

          {/* Storage Pool */}
          <div>
            <label htmlFor="storage" className="block text-sm font-medium text-text-primary mb-2">
              Storage Pool
            </label>
            <Select
              id="storage"
              value={formData.storage || 'local-lvm'}
              onChange={(e) => setFormData({ ...formData, storage: e.target.value })}
              options={[
                { value: 'local-lvm', label: 'local-lvm (LVM)' },
                { value: 'local-zfs', label: 'local-zfs (ZFS)' },
                { value: 'local', label: 'local (Directory)' },
              ]}
            />
            <p className="mt-1 text-sm text-text-muted">
              Choose the storage backend for your volume
            </p>
          </div>

          {/* Volume Type */}
          <div>
            <label htmlFor="type" className="block text-sm font-medium text-text-primary mb-2">
              Volume Type
            </label>
            <Select
              id="type"
              value={formData.type || 'ssd'}
              onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              options={[
                { value: 'ssd', label: 'SSD (High Performance)' },
                { value: 'hdd', label: 'HDD (Cost Effective)' },
              ]}
            />
            <p className="mt-1 text-sm text-text-muted">
              SSD provides faster I/O, HDD is more cost-effective for large data
            </p>
          </div>

          {/* Info Card */}
          <Card className="bg-info/5 border-info/20">
            <div className="flex items-start gap-3">
              <HardDrive className="w-5 h-5 text-info flex-shrink-0 mt-0.5" />
              <div className="space-y-2 text-sm">
                <p className="text-text-primary font-medium">About Persistent Volumes</p>
                <ul className="text-text-muted space-y-1 list-disc list-inside">
                  <li>Volumes persist independently of containers</li>
                  <li>Can be attached to running containers</li>
                  <li>Data remains intact when detached or container is deleted</li>
                  <li>Supports snapshots for backup and recovery</li>
                </ul>
              </div>
            </div>
          </Card>

          {/* Actions */}
          <div className="flex items-center gap-4 pt-4">
            <Button type="submit" disabled={loading}>
              {loading ? 'Creating...' : 'Create Volume'}
            </Button>
            <Link href="/volumes">
              <Button type="button" variant="ghost" disabled={loading}>
                Cancel
              </Button>
            </Link>
          </div>
        </form>
      </Card>
    </div>
  );
}
