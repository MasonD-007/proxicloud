'use client';

import React, { useEffect, useState, useRef } from 'react';
import { getTemplates, uploadTemplate } from '@/lib/api';
import Card from '@/components/ui/Card';
import Badge from '@/components/ui/Badge';
import Button from '@/components/ui/Button';
import type { Template } from '@/lib/types';

export default function TemplatesPage() {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [uploadSuccess, setUploadSuccess] = useState<string | null>(null);
  const [selectedStorage, setSelectedStorage] = useState('local');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const fetchTemplates = async () => {
    try {
      setLoading(true);
      const data = await getTemplates();
      setTemplates(data);
      setError(null);
    } catch (err) {
      console.error('Failed to fetch templates:', err);
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTemplates();
  }, []);

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Validate file extension
    const validExtensions = ['.tar.gz', '.tar.xz', '.tar.zst', '.tar.bz2', '.tgz'];
    const isValid = validExtensions.some(ext => file.name.toLowerCase().endsWith(ext));
    
    if (!isValid) {
      setUploadError('Invalid file format. Must be .tar.gz, .tar.xz, .tar.zst, .tar.bz2, or .tgz');
      return;
    }

    // Reset states
    setUploadError(null);
    setUploadSuccess(null);
    setUploading(true);
    setUploadProgress(0);

    try {
      const result = await uploadTemplate(file, selectedStorage, (progress) => {
        setUploadProgress(progress);
      });

      setUploadSuccess(`Template "${result.filename}" uploaded successfully!`);
      setUploadProgress(100);
      
      // Refresh template list after upload
      setTimeout(() => {
        fetchTemplates();
      }, 1000);

      // Clear success message after 5 seconds
      setTimeout(() => {
        setUploadSuccess(null);
        setUploadProgress(0);
      }, 5000);
    } catch (err) {
      console.error('Upload failed:', err);
      setUploadError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
  };

  const getTemplateName = (volid: string): string => {
    // Extract template name from volid
    // Format: storage:vztmpl/template-name.tar.zst
    const parts = volid.split('/');
    if (parts.length > 1) {
      const filename = parts[parts.length - 1];
      // Remove .tar.zst, .tar.gz, etc.
      return filename.replace(/\.(tar\.(zst|gz|xz|bz2)|tgz)$/, '');
    }
    return volid;
  };

  const getOS = (name: string): string => {
    const lowerName = name.toLowerCase();
    if (lowerName.includes('debian')) return 'Debian';
    if (lowerName.includes('ubuntu')) return 'Ubuntu';
    if (lowerName.includes('alpine')) return 'Alpine';
    if (lowerName.includes('centos')) return 'CentOS';
    if (lowerName.includes('rocky')) return 'Rocky Linux';
    if (lowerName.includes('fedora')) return 'Fedora';
    if (lowerName.includes('arch')) return 'Arch Linux';
    return 'Linux';
  };

  const getOSColor = (os: string): 'success' | 'warning' | 'error' | 'info' | 'default' => {
    const lowerOS = os.toLowerCase();
    if (lowerOS.includes('debian')) return 'error';
    if (lowerOS.includes('ubuntu')) return 'warning';
    if (lowerOS.includes('alpine')) return 'info';
    if (lowerOS.includes('centos') || lowerOS.includes('rocky')) return 'success';
    if (lowerOS.includes('fedora')) return 'info';
    return 'default';
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Templates</h1>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => (
            <Card key={i} className="p-6">
              <div className="animate-pulse space-y-4">
                <div className="h-4 bg-gray-700 rounded w-3/4"></div>
                <div className="h-3 bg-gray-700 rounded w-1/2"></div>
                <div className="h-3 bg-gray-700 rounded w-full"></div>
              </div>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-white">Templates</h1>
        </div>
        <Card className="p-6">
          <p className="text-red-400">Error loading templates: {error}</p>
          <p className="text-sm text-gray-400 mt-2">
            Make sure you have templates downloaded in Proxmox. You can download them from:
            Datacenter → Storage → local → CT Templates → Templates
          </p>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Templates</h1>
          <p className="text-gray-400 mt-1">
            {templates.length} LXC template{templates.length !== 1 ? 's' : ''} available
          </p>
        </div>
        <Button
          variant="primary"
          onClick={handleUploadClick}
          disabled={uploading}
        >
          <svg className="w-5 h-5 mr-2 inline" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          Upload Template
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept=".tar.gz,.tar.xz,.tar.zst,.tar.bz2,.tgz"
          onChange={handleFileSelect}
          className="hidden"
        />
      </div>

      {/* Upload Progress */}
      {uploading && (
        <Card className="p-4 bg-blue-900/20 border-blue-700/50">
          <div className="flex items-center justify-between mb-2">
            <span className="text-blue-300 font-medium">Uploading template...</span>
            <span className="text-blue-400 text-sm">{Math.round(uploadProgress)}%</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-2">
            <div
              className="bg-blue-500 h-2 rounded-full transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            ></div>
          </div>
        </Card>
      )}

      {/* Upload Success */}
      {uploadSuccess && (
        <Card className="p-4 bg-green-900/20 border-green-700/50">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-green-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div className="flex-1">
              <h3 className="text-sm font-semibold text-green-300">Upload Successful</h3>
              <p className="text-sm text-green-200">{uploadSuccess}</p>
            </div>
          </div>
        </Card>
      )}

      {/* Upload Error */}
      {uploadError && (
        <Card className="p-4 bg-red-900/20 border-red-700/50">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-red-400 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div className="flex-1">
              <h3 className="text-sm font-semibold text-red-300">Upload Failed</h3>
              <p className="text-sm text-red-200">{uploadError}</p>
            </div>
          </div>
        </Card>
      )}

      {/* Storage Selector */}
      <Card className="p-4">
        <div className="flex items-center gap-4">
          <label htmlFor="storage" className="text-sm font-medium text-gray-300">
            Upload to storage:
          </label>
          <select
            id="storage"
            value={selectedStorage}
            onChange={(e) => setSelectedStorage(e.target.value)}
            className="px-3 py-1.5 bg-gray-700 border border-gray-600 rounded text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={uploading}
          >
            <option value="local">local</option>
            <option value="local-lvm">local-lvm</option>
          </select>
          <span className="text-xs text-gray-400">
            Accepted formats: .tar.gz, .tar.xz, .tar.zst, .tar.bz2, .tgz
          </span>
        </div>
      </Card>

      {/* Templates Grid */}
      {templates.length === 0 ? (
        <Card className="p-6">
          <h2 className="text-xl font-semibold text-white mb-2">No Templates Available</h2>
          <p className="text-gray-400 mb-4">
            No LXC templates found. You can upload a template using the button above, or download templates in Proxmox.
          </p>
          <div className="bg-blue-900/20 border border-blue-700/50 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-blue-300 mb-2">How to download templates in Proxmox:</h3>
            <ol className="list-decimal list-inside text-sm text-blue-200 space-y-1">
              <li>Open Proxmox web interface</li>
              <li>Navigate to: Datacenter → Storage → local</li>
              <li>Click on &quot;CT Templates&quot; tab</li>
              <li>Click &quot;Templates&quot; button</li>
              <li>Select and download your desired templates (e.g., Debian, Ubuntu)</li>
            </ol>
          </div>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {templates.map((template) => {
            const name = getTemplateName(template.volid);
            const os = getOS(name);
            const color = getOSColor(os);

            return (
              <Card key={template.volid} className="p-6 hover:border-blue-600 transition-colors">
                <div className="flex items-start justify-between mb-4">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-white mb-2 break-all">
                      {name}
                    </h3>
                    <Badge variant={color}>{os}</Badge>
                  </div>
                </div>

                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-400">Format:</span>
                    <span className="text-gray-300 font-mono">{template.format}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Size:</span>
                    <span className="text-gray-300">{formatBytes(template.size)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Storage ID:</span>
                    <span className="text-gray-300 font-mono text-xs truncate max-w-[150px]" title={template.volid}>
                      {template.volid.split(':')[0]}
                    </span>
                  </div>
                </div>

                <div className="mt-4 pt-4 border-t border-gray-700">
                  <p className="text-xs text-gray-500 break-all">
                    {template.volid}
                  </p>
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {/* Info Card */}
      <Card className="p-4 bg-blue-900/20 border-blue-700/50">
        <div className="flex items-start gap-3">
          <div className="text-blue-400 mt-0.5">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-blue-300 mb-1">About Templates</h3>
            <p className="text-sm text-blue-200">
              These are LXC container templates available on your Proxmox node. 
              You can use them to quickly create new containers with pre-configured operating systems.
              To use a template, go to the Containers page and click &quot;Create Container&quot;.
            </p>
          </div>
        </div>
      </Card>
    </div>
  );
}
