'use client';

import React from 'react';
import Button from './ui/Button';
import { formatErrorMessage, getErrorSuggestion, isRetryable } from '@/lib/errors';

interface ErrorDisplayProps {
  error: unknown;
  onRetry?: () => void;
  onDismiss?: () => void;
  showDetails?: boolean;
}

export function ErrorDisplay({ 
  error, 
  onRetry, 
  onDismiss,
  showDetails = false 
}: ErrorDisplayProps) {
  const [showFullDetails, setShowFullDetails] = React.useState(false);
  const message = formatErrorMessage(error);
  const suggestion = getErrorSuggestion(error);
  const canRetry = isRetryable(error);

  return (
    <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4">
      {/* Error Icon and Message */}
      <div className="flex items-start gap-3">
        <div className="flex-shrink-0">
          <svg
            className="w-5 h-5 text-red-500 mt-0.5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        </div>

        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-red-400 mb-1">
            Error
          </h3>
          <p className="text-sm text-gray-300">
            {message}
          </p>

          {/* Suggestion */}
          {suggestion && (
            <p className="text-sm text-gray-400 mt-2">
              <span className="font-medium">Suggestion:</span> {suggestion}
            </p>
          )}

          {/* Error Details Toggle */}
          {showDetails && error instanceof Error && (
            <div className="mt-3">
              <button
                onClick={() => setShowFullDetails(!showFullDetails)}
                className="text-xs text-gray-500 hover:text-gray-400"
              >
                {showFullDetails ? '▼' : '▶'} {showFullDetails ? 'Hide' : 'Show'} technical details
              </button>

              {showFullDetails && (
                <div className="mt-2 bg-[#0f1419] rounded p-2 overflow-auto">
                  <pre className="text-xs text-gray-500 whitespace-pre-wrap break-words">
                    {error.stack || error.message}
                  </pre>
                </div>
              )}
            </div>
          )}

          {/* Actions */}
          {(onRetry || onDismiss) && (
            <div className="flex gap-2 mt-3">
              {onRetry && canRetry && (
                <Button
                  onClick={onRetry}
                  size="sm"
                  variant="secondary"
                >
                  Try Again
                </Button>
              )}
              {onDismiss && (
                <Button
                  onClick={onDismiss}
                  size="sm"
                  variant="secondary"
                >
                  Dismiss
                </Button>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Inline error for form fields
interface FieldErrorProps {
  error?: string;
}

export function FieldError({ error }: FieldErrorProps) {
  if (!error) return null;

  return (
    <p className="text-sm text-red-400 mt-1 flex items-center gap-1">
      <svg
        className="w-4 h-4 flex-shrink-0"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        />
      </svg>
      {error}
    </p>
  );
}

// Empty state with error message
interface EmptyStateErrorProps {
  title: string;
  message: string;
  onRetry?: () => void;
}

export function EmptyStateError({ title, message, onRetry }: EmptyStateErrorProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 px-4">
      <div className="flex items-center justify-center w-16 h-16 bg-red-500/10 rounded-full mb-4">
        <svg
          className="w-8 h-8 text-red-500"
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
      
      <h3 className="text-lg font-semibold text-white mb-2">
        {title}
      </h3>
      
      <p className="text-gray-400 text-center max-w-md mb-4">
        {message}
      </p>

      {onRetry && (
        <Button onClick={onRetry} variant="primary">
          Try Again
        </Button>
      )}
    </div>
  );
}
