import { APIError } from './api';

// Error message formatting
export function formatErrorMessage(error: unknown): string {
  if (error instanceof APIError) {
    // Format API errors with context
    if (error.status === 0) {
      return 'Unable to connect to the server. Please check your connection.';
    }
    
    if (error.status === 401) {
      return 'Authentication failed. Please check your Proxmox credentials.';
    }
    
    if (error.status === 403) {
      return 'Permission denied. Your API token may not have sufficient privileges.';
    }
    
    if (error.status === 404) {
      return 'Resource not found. It may have been deleted or moved.';
    }
    
    if (error.status === 429) {
      return 'Too many requests. Please wait a moment and try again.';
    }
    
    if (error.status && error.status >= 500) {
      return `Server error (${error.status}). The Proxmox server may be experiencing issues.`;
    }
    
    return error.message || 'An unknown error occurred.';
  }
  
  if (error instanceof Error) {
    return error.message;
  }
  
  return 'An unknown error occurred.';
}

// User-friendly action suggestions based on error type
export function getErrorSuggestion(error: unknown): string | null {
  if (error instanceof APIError) {
    if (error.status === 0) {
      return 'Check that the backend is running and accessible at the configured URL.';
    }
    
    if (error.status === 401 || error.status === 403) {
      return 'Verify your Proxmox API token has the correct permissions and has not expired.';
    }
    
    if (error.status === 404) {
      return 'Try refreshing the page or navigating back to the main dashboard.';
    }
    
    if (error.status === 429) {
      return 'Wait a few seconds before making another request.';
    }
    
    if (error.status && error.status >= 500) {
      return 'Check the Proxmox server status and logs. The issue may be temporary.';
    }
  }
  
  return null;
}

// Determine if error should show a retry button
export function isRetryable(error: unknown): boolean {
  if (error instanceof APIError) {
    return error.isRetryable || false;
  }
  
  return false;
}

// Toast notification helper (for future use with a toast library)
export interface ToastOptions {
  title: string;
  description?: string;
  type: 'success' | 'error' | 'warning' | 'info';
  duration?: number;
}

export function createErrorToast(error: unknown): ToastOptions {
  return {
    title: 'Error',
    description: formatErrorMessage(error),
    type: 'error',
    duration: 5000,
  };
}

export function createSuccessToast(message: string): ToastOptions {
  return {
    title: 'Success',
    description: message,
    type: 'success',
    duration: 3000,
  };
}

// Error logging utility
export function logError(error: unknown, context?: Record<string, unknown>) {
  console.error('Application Error:', {
    error,
    message: error instanceof Error ? error.message : 'Unknown error',
    stack: error instanceof Error ? error.stack : undefined,
    context,
    timestamp: new Date().toISOString(),
  });
  
  // Could send to external logging service here
  if (typeof window !== 'undefined' && process.env.NODE_ENV === 'production') {
    try {
      // Store recent errors for debugging
      const recentErrors = JSON.parse(localStorage.getItem('recentErrors') || '[]');
      recentErrors.push({
        message: error instanceof Error ? error.message : String(error),
        timestamp: new Date().toISOString(),
        context,
      });
      
      // Keep only last 10 errors
      if (recentErrors.length > 10) {
        recentErrors.shift();
      }
      
      localStorage.setItem('recentErrors', JSON.stringify(recentErrors));
    } catch {
      // Ignore localStorage errors
    }
  }
}

// React hook-friendly error handler
export function handleError(error: unknown, context?: Record<string, unknown>) {
  logError(error, context);
  return formatErrorMessage(error);
}
