'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Plus, Trash2, FolderOpen, Box } from 'lucide-react';
import Card from '@/components/ui/Card';
import Button from '@/components/ui/Button';
import Badge from '@/components/ui/Badge';
import { getProjects, deleteProject, getProjectContainers } from '@/lib/api';
import type { Project } from '@/lib/types';

interface ProjectWithCount extends Project {
  containerCount?: number;
}

export default function ProjectsPage() {
  const [projects, setProjects] = useState<ProjectWithCount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    loadProjects();
  }, []);

  async function loadProjects() {
    try {
      setLoading(true);
      setError(null);
      const data = await getProjects();
      
      // Fetch container counts for each project
      const projectsWithCounts = await Promise.all(
        data.map(async (project) => {
          try {
            const containers = await getProjectContainers(project.id);
            return { ...project, containerCount: containers.length };
          } catch {
            return { ...project, containerCount: 0 };
          }
        })
      );
      
      setProjects(projectsWithCounts);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Are you sure you want to delete project "${name}"? This will only work if no containers are assigned to it.`)) {
      return;
    }

    try {
      setActionLoading(id);
      await deleteProject(id);
      await loadProjects();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete project');
    } finally {
      setActionLoading(null);
    }
  }

  function formatDate(timestamp: number): string {
    return new Date(timestamp * 1000).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-text-secondary">Loading projects...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="text-error mb-2">Error loading projects</div>
          <div className="text-text-muted text-sm">{error}</div>
          <Button onClick={loadProjects} className="mt-4" size="sm">
            Retry
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-text-primary">Projects</h1>
        <Link href="/projects/create">
          <Button>
            <Plus className="w-4 h-4 mr-2" />
            Create Project
          </Button>
        </Link>
      </div>

      {projects.length === 0 ? (
        <Card>
          <div className="text-center py-12">
            <FolderOpen className="w-16 h-16 text-text-muted mx-auto mb-4" />
            <h2 className="text-xl font-semibold text-text-primary mb-2">No projects yet</h2>
            <p className="text-text-secondary mb-6">
              Create your first project to organize containers by application or environment
            </p>
            <Link href="/projects/create">
              <Button>
                <Plus className="w-4 h-4 mr-2" />
                Create Project
              </Button>
            </Link>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <Card key={project.id} className="hover:border-primary-dark transition-colors">
              <Link href={`/projects/${project.id}`}>
                <div className="space-y-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-text-primary mb-1">
                        {project.name}
                      </h3>
                      {project.description && (
                        <p className="text-sm text-text-secondary line-clamp-2">
                          {project.description}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-2 ml-4">
                      <Box className="w-5 h-5 text-primary" />
                      <span className="text-lg font-bold text-text-primary">
                        {project.containerCount ?? 0}
                      </span>
                    </div>
                  </div>

              {project.tags && project.tags.length > 0 && (
                <div className="flex flex-wrap gap-2 mt-3">
                  {project.tags.map((tag) => (
                    <Badge key={tag} variant="default">
                      {tag}
                    </Badge>
                  ))}
                </div>
              )}

                  <div className="flex items-center justify-between pt-4 border-t border-border">
                    <span className="text-xs text-text-muted">
                      Created {formatDate(project.created_at)}
                    </span>
                    <button
                      onClick={(e) => {
                        e.preventDefault();
                        handleDelete(project.id, project.name);
                      }}
                      disabled={actionLoading === project.id}
                      className="text-error hover:text-error/80 disabled:opacity-50"
                      title="Delete project"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </Link>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
