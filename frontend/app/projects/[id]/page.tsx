'use client';

import { useEffect, useState, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft, Trash2, Play, Square, RotateCw, Plus, Network } from 'lucide-react';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getProject, getProjectContainers, deleteProject, startContainer, stopContainer, rebootContainer } from '@/lib/api';
import { formatBytes, formatCPU, formatUptime } from '@/lib/utils';
import type { Project, Container } from '@/lib/types';

export default function ProjectDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const projectId = params.id as string;
  
  const [project, setProject] = useState<Project | null>(null);
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<number | null>(null);

  const loadProjectData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [projectData, projectContainersData] = await Promise.all([
        getProject(projectId),
        getProjectContainers(projectId),
      ]);
      setProject(projectData);
      // Extract containers array from the response object
      setContainers(projectContainersData.containers || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load project');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    loadProjectData();
  }, [loadProjectData]);

  async function handleDelete() {
    if (!project) return;
    
    if (!confirm(`Are you sure you want to delete project "${project.name}"? This will only work if no containers are assigned to it.`)) {
      return;
    }

    try {
      await deleteProject(projectId);
      router.push('/projects');
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete project');
    }
  }

  async function handleContainerAction(vmid: number, action: 'start' | 'stop' | 'reboot') {
    try {
      setActionLoading(vmid);
      switch (action) {
        case 'start':
          await startContainer(vmid);
          break;
        case 'stop':
          await stopContainer(vmid);
          break;
        case 'reboot':
          await rebootContainer(vmid);
          break;
      }
      await loadProjectData();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setActionLoading(null);
    }
  }

  function formatDate(timestamp: number): string {
    return new Date(timestamp * 1000).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  function calculateTotals() {
    if (!Array.isArray(containers) || containers.length === 0) {
      return { cpu: 0, memory: 0, maxMemory: 0, running: 0, stopped: 0 };
    }
    
    return containers.reduce(
      (acc, container) => ({
        cpu: acc.cpu + (container.cpu || 0),
        memory: acc.memory + (container.mem || 0),
        maxMemory: acc.maxMemory + (container.maxmem || 0),
        running: acc.running + (container.status === 'running' ? 1 : 0),
        stopped: acc.stopped + (container.status === 'stopped' ? 1 : 0),
      }),
      { cpu: 0, memory: 0, maxMemory: 0, running: 0, stopped: 0 }
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading project...</div>
      </div>
    );
  }

  if (error || !project) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading project</div>
          <div className="text-text-muted text-sm">{error}</div>
          <div className="flex gap-4 justify-center mt-4">
            <Button onClick={loadProjectData} size="sm">
              Retry
            </Button>
            <Link href="/projects">
              <Button variant="outline" size="sm">
                Back to Projects
              </Button>
            </Link>
          </div>
        </div>
      </div>
    );
  }

  const totals = calculateTotals();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <Link href="/projects">
            <Button variant="outline" size="sm">
              <ArrowLeft className="w-4 h-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold text-text-primary mb-2">{project.name}</h1>
            {project.description && (
              <p className="text-text-secondary">{project.description}</p>
            )}
                {project.tags && project.tags.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {project.tags.map((tag) => (
                      <Badge key={tag} variant="default">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                )}
          </div>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={handleDelete}>
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <div className="space-y-1">
            <div className="text-sm text-text-secondary">Total Containers</div>
            <div className="text-2xl font-bold text-text-primary">{containers.length}</div>
          </div>
        </Card>
        <Card>
          <div className="space-y-1">
            <div className="text-sm text-text-secondary">Running</div>
            <div className="text-2xl font-bold text-success">{totals.running}</div>
          </div>
        </Card>
        <Card>
          <div className="space-y-1">
            <div className="text-sm text-text-secondary">Total CPU</div>
            <div className="text-2xl font-bold text-text-primary">{formatCPU(totals.cpu)}</div>
          </div>
        </Card>
        <Card>
          <div className="space-y-1">
            <div className="text-sm text-text-secondary">Total RAM</div>
            <div className="text-2xl font-bold text-text-primary">
              {formatBytes(totals.memory)} / {formatBytes(totals.maxMemory)}
            </div>
          </div>
        </Card>
      </div>

      {/* Network Configuration */}
      {project.network && (
        <Card>
          <div className="flex items-center gap-2 mb-4">
            <Network className="w-5 h-5 text-primary" />
            <h2 className="text-xl font-semibold text-text-primary">Network Configuration</h2>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {project.network.subnet && (
              <div>
                <div className="text-sm text-text-secondary mb-1">Subnet</div>
                <div className="text-text-primary font-mono">{project.network.subnet}</div>
              </div>
            )}
            {project.network.gateway && (
              <div>
                <div className="text-sm text-text-secondary mb-1">Gateway</div>
                <div className="text-text-primary font-mono">{project.network.gateway}</div>
              </div>
            )}
            {project.network.nameserver && (
              <div>
                <div className="text-sm text-text-secondary mb-1">DNS Nameserver</div>
                <div className="text-text-primary font-mono">{project.network.nameserver}</div>
              </div>
            )}
            {project.network.vlan_tag && (
              <div>
                <div className="text-sm text-text-secondary mb-1">VLAN Tag</div>
                <div className="text-text-primary font-mono">{project.network.vlan_tag}</div>
              </div>
            )}
          </div>
          <div className="mt-4 p-3 bg-background-elevated rounded-lg">
            <p className="text-sm text-text-muted">
              <span className="text-primary font-medium">Note:</span> New containers created in this project will automatically use these network settings.
            </p>
          </div>
        </Card>
      )}

      {/* Containers List */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-text-primary">Containers</h2>
          <Link href={`/containers/create?project_id=${projectId}`}>
            <Button size="sm">
              <Plus className="w-4 h-4 mr-2" />
              Add Container
            </Button>
          </Link>
        </div>

        {containers.length === 0 ? (
          <div className="text-center py-8 text-text-secondary">
            No containers assigned to this project yet
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">Name</th>
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">VMID</th>
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">Status</th>
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">CPU</th>
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">Memory</th>
                  <th className="text-left py-3 px-4 font-medium text-text-secondary">Uptime</th>
                  <th className="text-right py-3 px-4 font-medium text-text-secondary">Actions</th>
                </tr>
              </thead>
              <tbody>
                {containers.map((container) => (
                  <tr key={container.vmid} className="border-b border-border last:border-0">
                    <td className="py-3 px-4">
                      <Link
                        href={`/containers/${container.vmid}`}
                        className="text-text-primary hover:text-primary font-medium"
                      >
                        {container.name}
                      </Link>
                    </td>
                    <td className="py-3 px-4 text-text-secondary">{container.vmid}</td>
                    <td className="py-3 px-4">
                      <Badge variant={container.status === 'running' ? 'success' : 'default'}>
                        {container.status}
                      </Badge>
                    </td>
                    <td className="py-3 px-4 text-text-secondary">{formatCPU(container.cpu)}</td>
                    <td className="py-3 px-4 text-text-secondary">
                      {formatBytes(container.mem)} / {formatBytes(container.maxmem)}
                    </td>
                    <td className="py-3 px-4 text-text-secondary">
                      {container.status === 'running' ? formatUptime(container.uptime) : '-'}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center justify-end gap-2">
                        {container.status === 'stopped' ? (
                          <button
                            onClick={() => handleContainerAction(container.vmid, 'start')}
                            disabled={actionLoading === container.vmid}
                            className="text-success hover:text-success/80 disabled:opacity-50"
                            title="Start"
                          >
                            <Play className="w-4 h-4" />
                          </button>
                        ) : (
                          <>
                            <button
                              onClick={() => handleContainerAction(container.vmid, 'reboot')}
                              disabled={actionLoading === container.vmid}
                              className="text-warning hover:text-warning/80 disabled:opacity-50"
                              title="Reboot"
                            >
                              <RotateCw className="w-4 h-4" />
                            </button>
                            <button
                              onClick={() => handleContainerAction(container.vmid, 'stop')}
                              disabled={actionLoading === container.vmid}
                              className="text-error hover:text-error/80 disabled:opacity-50"
                              title="Stop"
                            >
                              <Square className="w-4 h-4" />
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Project Metadata */}
      <Card>
        <h2 className="text-xl font-semibold text-text-primary mb-4">Project Information</h2>
        <dl className="space-y-2 text-sm">
          <div className="flex justify-between">
            <dt className="text-text-secondary">Project ID</dt>
            <dd className="text-text-primary font-mono">{project.id}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-text-secondary">Created</dt>
            <dd className="text-text-primary">{formatDate(project.created_at)}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-text-secondary">Last Updated</dt>
            <dd className="text-text-primary">{formatDate(project.updated_at)}</dd>
          </div>
        </dl>
      </Card>
    </div>
  );
}
