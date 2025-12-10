'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Server, Activity, HardDrive, Plus, FolderOpen } from 'lucide-react';
import Card, { CardHeader, CardTitle, CardContent } from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getDashboard, getContainers, getProjects } from '@/lib/api';
import { formatBytes, formatCPU } from '@/lib/utils';
import type { DashboardStats, Container, Project } from '@/lib/types';

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [containers, setContainers] = useState<Container[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboard();
  }, []);

  async function loadDashboard() {
    try {
      setLoading(true);
      setError(null);
      const [dashboardData, containersData, projectsData] = await Promise.all([
        getDashboard(),
        getContainers(),
        getProjects(),
      ]);
      setStats(dashboardData);
      setContainers(containersData.slice(0, 5)); // Recent 5
      setProjects(projectsData.slice(0, 5)); // Recent 5
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard');
    } finally {
      setLoading(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading dashboard</div>
          <div className="text-text-muted text-sm">{error}</div>
          <Button onClick={loadDashboard} className="mt-4" size="sm">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-text-primary">Dashboard</h1>
        <Link href="/containers/create">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Container
          </Button>
        </Link>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-muted text-sm mb-1">Total Projects</div>
              <div className="text-3xl font-bold text-text-primary">
                {projects.length}
              </div>
              <div className="text-sm text-text-muted mt-1">
                Active projects
              </div>
            </div>
            <div className="p-3 bg-success/10 rounded-lg">
              <FolderOpen className="w-6 h-6 text-success" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-muted text-sm mb-1">Total Containers</div>
              <div className="text-3xl font-bold text-text-primary">
                {stats?.total_containers || 0}
              </div>
              <div className="text-sm text-success mt-1">
                {stats?.running_containers || 0} running
              </div>
            </div>
            <div className="p-3 bg-primary/10 rounded-lg">
              <Server className="w-6 h-6 text-primary" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-muted text-sm mb-1">CPU Usage</div>
              <div className="text-3xl font-bold text-text-primary">
                {formatCPU(stats?.total_cpu || 0)}
              </div>
              <div className="text-sm text-text-muted mt-1">Total CPU</div>
            </div>
            <div className="p-3 bg-info/10 rounded-lg">
              <Activity className="w-6 h-6 text-info" />
            </div>
          </div>
        </Card>

        <Card>
          <div className="flex items-center justify-between">
            <div>
              <div className="text-text-muted text-sm mb-1">Memory</div>
              <div className="text-3xl font-bold text-text-primary">
                {formatBytes(stats?.used_memory || 0)}
              </div>
              <div className="text-sm text-text-muted mt-1">
                of {formatBytes(stats?.total_memory || 0)}
              </div>
            </div>
            <div className="p-3 bg-warning/10 rounded-lg">
              <HardDrive className="w-6 h-6 text-warning" />
            </div>
          </div>
        </Card>
      </div>

      {/* Recent Projects */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Recent Projects</CardTitle>
            <Link href="/projects">
              <Button variant="ghost" size="sm">
                View All
              </Button>
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          {projects.length === 0 ? (
            <div className="text-center py-8 text-text-muted">
              No projects found
            </div>
          ) : (
            <div className="space-y-2">
              {projects.map((project) => (
                <Link key={project.id} href={`/projects/${project.id}`}>
                  <div className="flex items-center justify-between p-3 rounded-lg hover:bg-surface-elevated transition-colors">
                    <div className="flex items-center gap-3">
                      <FolderOpen className="w-5 h-5 text-text-muted" />
                      <div>
                        <div className="font-medium text-text-primary">
                          {project.name}
                        </div>
                        {project.description && (
                          <div className="text-sm text-text-muted line-clamp-1">
                            {project.description}
                          </div>
                        )}
                      </div>
                    </div>
                    <Badge variant="default">
                      {project.container_count} containers
                    </Badge>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Recent Containers */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Recent Containers</CardTitle>
            <Link href="/containers">
              <Button variant="ghost" size="sm">
                View All
              </Button>
            </Link>
          </div>
        </CardHeader>
        <CardContent>
          {containers.length === 0 ? (
            <div className="text-center py-8 text-text-muted">
              No containers found
            </div>
          ) : (
            <div className="space-y-2">
              {containers.map((container) => (
                <Link key={container.vmid} href={`/containers/${container.vmid}`}>
                  <div className="flex items-center justify-between p-3 rounded-lg hover:bg-surface-elevated transition-colors">
                    <div className="flex items-center gap-3">
                      <Server className="w-5 h-5 text-text-muted" />
                      <div>
                        <div className="font-medium text-text-primary">
                          {container.name}
                        </div>
                        <div className="text-sm text-text-muted">
                          ID: {container.vmid}
                        </div>
                      </div>
                    </div>
                    <Badge variant={container.status === 'running' ? 'success' : 'default'}>
                      {container.status}
                    </Badge>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
