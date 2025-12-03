'use client';

import React from 'react';
import Button from '@/components/ui/Button';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<{ error: Error; reset: () => void }>;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    
    // Log to external service if configured
    if (typeof window !== 'undefined') {
      // Could send to error tracking service here
      try {
        localStorage.setItem('lastError', JSON.stringify({
          error: error.message,
          stack: error.stack,
          timestamp: new Date().toISOString(),
        }));
      } catch (e) {
        console.error('Failed to log error:', e);
      }
    }
  }

  reset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) {
        const FallbackComponent = this.props.fallback;
        return <FallbackComponent error={this.state.error} reset={this.reset} />;
      }

      return (
        <DefaultErrorFallback error={this.state.error} reset={this.reset} />
      );
    }

    return this.props.children;
  }
}

// Default error fallback component
function DefaultErrorFallback({ error, reset }: { error: Error; reset: () => void }) {
  return (
    <div className="min-h-screen bg-[#0f1419] flex items-center justify-center px-4">
      <div className="max-w-md w-full">
        <div className="bg-[#1a1f2e] border border-red-500/30 rounded-lg p-6">
          {/* Error Icon */}
          <div className="flex items-center justify-center w-12 h-12 bg-red-500/10 rounded-full mb-4">
            <svg
              className="w-6 h-6 text-red-500"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>

          {/* Error Message */}
          <h2 className="text-xl font-semibold text-white mb-2">
            Something went wrong
          </h2>
          <p className="text-gray-400 mb-4">
            An unexpected error occurred. You can try refreshing the page or going back.
          </p>

          {/* Error Details (collapsed by default) */}
          <details className="mb-4">
            <summary className="text-sm text-gray-500 cursor-pointer hover:text-gray-400 mb-2">
              Show error details
            </summary>
            <div className="bg-[#0f1419] rounded p-3 overflow-auto">
              <p className="text-xs text-red-400 font-mono break-words">
                {error.message}
              </p>
              {error.stack && (
                <pre className="text-xs text-gray-500 mt-2 overflow-x-auto">
                  {error.stack}
                </pre>
              )}
            </div>
          </details>

          {/* Actions */}
          <div className="flex gap-2">
            <Button
              onClick={reset}
              variant="primary"
              className="flex-1"
            >
              Try Again
            </Button>
            <Button
              onClick={() => window.location.href = '/'}
              variant="secondary"
              className="flex-1"
            >
              Go Home
            </Button>
          </div>

          {/* Help Text */}
          <p className="text-xs text-gray-500 mt-4 text-center">
            If this problem persists, please check the browser console for more details.
          </p>
        </div>
      </div>
    </div>
  );
}

export default ErrorBoundary;
