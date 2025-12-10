import { Container, CreateContainerRequest, DashboardStats, MetricsData, MetricsSummary, Template, Volume, CreateVolumeRequest, AttachVolumeRequest, DetachVolumeRequest, Snapshot, CreateSnapshotRequest, RestoreSnapshotRequest, CloneSnapshotRequest, Project, CreateProjectRequest, UpdateProjectRequest, AssignProjectRequest, ProjectContainersResponse } from './types';

// Support runtime API URL configuration
// In standalone mode, this will be available at window location
const getAPIUrl = (): string => {
  // Check if we're in browser
  if (typeof window !== 'undefined') {
    // Use relative path when deployed (backend should be on same host)
    // This makes it work regardless of the host IP
    return `${window.location.protocol}//${window.location.hostname}:8080/api`;
  }
  
  // Server-side: use environment variable or default
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';
};

const API_URL = getAPIUrl();

// Global state for cache status
let isUsingCache = false;
let cacheListeners: Array<(cached: boolean) => void> = [];

export function onCacheStatusChange(listener: (cached: boolean) => void) {
  cacheListeners.push(listener);
  return () => {
    cacheListeners = cacheListeners.filter((l) => l !== listener);
  };
}

function notifyCacheStatus(cached: boolean) {
  if (isUsingCache !== cached) {
    isUsingCache = cached;
    cacheListeners.forEach((listener) => listener(cached));
  }
}

export function isCached(): boolean {
  return isUsingCache;
}

// Retry configuration
interface RetryOptions {
  maxRetries?: number;
  retryDelay?: number;
  retryOn?: number[]; // HTTP status codes to retry on
  backoff?: 'linear' | 'exponential';
}

const DEFAULT_RETRY_OPTIONS: RetryOptions = {
  maxRetries: 3,
  retryDelay: 1000, // 1 second
  retryOn: [408, 429, 500, 502, 503, 504], // Request timeout, rate limit, server errors
  backoff: 'exponential',
};

// Sleep utility
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// Calculate retry delay with backoff
function getRetryDelay(attempt: number, options: RetryOptions): number {
  const baseDelay = options.retryDelay || 1000;
  
  if (options.backoff === 'exponential') {
    return baseDelay * Math.pow(2, attempt);
  }
  
  return baseDelay * (attempt + 1);
}

// Enhanced error class
export class APIError extends Error {
  constructor(
    message: string,
    public status?: number,
    public statusText?: string,
    public endpoint?: string,
    public isRetryable?: boolean
  ) {
    super(message);
    this.name = 'APIError';
  }
}

async function fetchAPI<T>(
  endpoint: string,
  options?: RequestInit,
  retryOptions: RetryOptions = DEFAULT_RETRY_OPTIONS
): Promise<T> {
  const maxRetries = retryOptions.maxRetries || 3;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(`${API_URL}${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
        signal: AbortSignal.timeout(60000), // 60 second timeout for slow networks
        ...options,
      });

      // Check if response is from cache
      const cacheStatus = response.headers.get('X-Cache-Status');
      notifyCacheStatus(cacheStatus === 'HIT');

      // Check if we should retry based on status code
      const shouldRetry = retryOptions.retryOn?.includes(response.status);
      
      if (!response.ok) {
        const errorMessage = await response.text().catch(() => response.statusText);
        const error = new APIError(
          errorMessage || `API error: ${response.statusText}`,
          response.status,
          response.statusText,
          endpoint,
          shouldRetry
        );

        // Retry on specific status codes if not the last attempt
        if (shouldRetry && attempt < maxRetries) {
          lastError = error;
          const delay = getRetryDelay(attempt, retryOptions);
          console.warn(`API request failed (attempt ${attempt + 1}/${maxRetries + 1}), retrying in ${delay}ms...`);
          await sleep(delay);
          continue;
        }

        throw error;
      }

      return response.json();
    } catch (error) {
      // Network errors (e.g., offline, CORS)
      if (error instanceof TypeError && error.message.includes('fetch')) {
        const networkError = new APIError(
          'Network error: Unable to connect to the server',
          0,
          'Network Error',
          endpoint,
          true
        );

        // Retry network errors if not the last attempt
        if (attempt < maxRetries) {
          lastError = networkError;
          const delay = getRetryDelay(attempt, retryOptions);
          console.warn(`Network error (attempt ${attempt + 1}/${maxRetries + 1}), retrying in ${delay}ms...`);
          await sleep(delay);
          continue;
        }

        throw networkError;
      }

      // Re-throw APIError or other errors
      if (error instanceof APIError) {
        if (error.isRetryable && attempt < maxRetries) {
          lastError = error;
          const delay = getRetryDelay(attempt, retryOptions);
          console.warn(`API error (attempt ${attempt + 1}/${maxRetries + 1}), retrying in ${delay}ms...`);
          await sleep(delay);
          continue;
        }
      }

      throw error;
    }
  }

  // If we exhausted all retries, throw the last error
  throw lastError || new APIError('Request failed after all retry attempts', 0, 'Max Retries Exceeded', endpoint, false);
}

// Health check
export async function healthCheck(): Promise<{ status: string }> {
  return fetchAPI('/health');
}

// Dashboard
export async function getDashboard(): Promise<DashboardStats> {
  return fetchAPI('/dashboard');
}

// Containers
export async function getContainers(): Promise<Container[]> {
  return fetchAPI('/containers');
}

export async function getContainer(vmid: number): Promise<Container> {
  return fetchAPI(`/containers/${vmid}`);
}

export async function createContainer(data: CreateContainerRequest): Promise<{ vmid: number }> {
  return fetchAPI('/containers', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function startContainer(vmid: number): Promise<void> {
  await fetchAPI(`/containers/${vmid}/start`, { method: 'POST' });
}

export async function stopContainer(vmid: number): Promise<void> {
  await fetchAPI(`/containers/${vmid}/stop`, { method: 'POST' });
}

export async function rebootContainer(vmid: number): Promise<void> {
  await fetchAPI(`/containers/${vmid}/reboot`, { method: 'POST' });
}

export async function deleteContainer(vmid: number): Promise<void> {
  await fetchAPI(`/containers/${vmid}`, { method: 'DELETE' });
}

// Metrics
export async function getMetrics(vmid: number, timeframe: string): Promise<MetricsData[]> {
  return fetchAPI(`/containers/${vmid}/metrics?timeframe=${timeframe}`);
}

export async function getMetricsSummary(vmid: number, hours: number): Promise<MetricsSummary> {
  return fetchAPI(`/containers/${vmid}/metrics/summary?hours=${hours}`);
}

// Templates
export async function getTemplates(): Promise<Template[]> {
  return fetchAPI('/templates');
}

export async function uploadTemplate(file: File, storage: string = 'local', onProgress?: (progress: number) => void): Promise<{ status: string; filename: string; storage: string }> {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('storage', storage);

  // Use XMLHttpRequest for progress tracking
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();

    // Track upload progress
    if (onProgress) {
      xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
          const percentComplete = (e.loaded / e.total) * 100;
          onProgress(percentComplete);
        }
      });
    }

    // Handle completion
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const response = JSON.parse(xhr.responseText);
          resolve(response);
        } catch {
          reject(new APIError('Failed to parse upload response', xhr.status, xhr.statusText, '/templates/upload'));
        }
      } else {
        reject(new APIError(
          xhr.responseText || `Upload failed: ${xhr.statusText}`,
          xhr.status,
          xhr.statusText,
          '/templates/upload'
        ));
      }
    });

    // Handle errors
    xhr.addEventListener('error', () => {
      reject(new APIError('Network error during upload', 0, 'Network Error', '/templates/upload'));
    });

    xhr.addEventListener('abort', () => {
      reject(new APIError('Upload aborted', 0, 'Aborted', '/templates/upload'));
    });

    // Send request
    xhr.open('POST', `${API_URL}/templates/upload`);
    xhr.send(formData);
  });
}

// Volumes
export async function getVolumes(): Promise<Volume[]> {
  return fetchAPI('/volumes');
}

export async function getVolume(volid: string): Promise<Volume> {
  return fetchAPI(`/volumes/${encodeURIComponent(volid)}`);
}

export async function createVolume(data: CreateVolumeRequest): Promise<Volume> {
  return fetchAPI('/volumes', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function deleteVolume(volid: string): Promise<void> {
  await fetchAPI(`/volumes/${encodeURIComponent(volid)}`, { method: 'DELETE' });
}

export async function attachVolume(volid: string, vmid: number, data?: AttachVolumeRequest): Promise<void> {
  await fetchAPI(`/volumes/${encodeURIComponent(volid)}/attach/${vmid}`, {
    method: 'POST',
    body: data ? JSON.stringify(data) : JSON.stringify({ vmid }),
  });
}

export async function detachVolume(volid: string, vmid: number, data?: DetachVolumeRequest): Promise<void> {
  await fetchAPI(`/volumes/${encodeURIComponent(volid)}/detach/${vmid}`, {
    method: 'POST',
    body: data ? JSON.stringify(data) : JSON.stringify({ vmid }),
  });
}

// Snapshots
export async function getSnapshots(volid: string): Promise<Snapshot[]> {
  return fetchAPI(`/volumes/${encodeURIComponent(volid)}/snapshots`);
}

export async function createSnapshot(volid: string, data: CreateSnapshotRequest): Promise<Snapshot> {
  return fetchAPI(`/volumes/${encodeURIComponent(volid)}/snapshots`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function restoreSnapshot(volid: string, data: RestoreSnapshotRequest): Promise<void> {
  await fetchAPI(`/volumes/${encodeURIComponent(volid)}/snapshots/restore`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function cloneSnapshot(volid: string, data: CloneSnapshotRequest): Promise<Volume> {
  return fetchAPI(`/volumes/${encodeURIComponent(volid)}/snapshots/clone`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// Projects
export async function getProjects(): Promise<Project[]> {
  return fetchAPI('/projects');
}

export async function getProject(id: string): Promise<Project> {
  return fetchAPI(`/projects/${id}`);
}

export async function createProject(data: CreateProjectRequest): Promise<Project> {
  return fetchAPI('/projects', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function updateProject(id: string, data: UpdateProjectRequest): Promise<Project> {
  return fetchAPI(`/projects/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export async function deleteProject(id: string): Promise<void> {
  await fetchAPI(`/projects/${id}`, { method: 'DELETE' });
}

export async function getProjectContainers(id: string): Promise<ProjectContainersResponse> {
  return fetchAPI(`/projects/${id}/containers`);
}

export async function assignContainerProject(vmid: number, data: AssignProjectRequest): Promise<void> {
  await fetchAPI(`/containers/${vmid}/project`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

