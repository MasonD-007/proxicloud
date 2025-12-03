'use client';

import { useEffect, useState } from 'react';
import { AlertCircle, WifiOff } from 'lucide-react';
import { onCacheStatusChange } from '@/lib/api';

export default function OfflineBanner() {
  const [isOffline, setIsOffline] = useState(false);

  useEffect(() => {
    // Listen for cache status changes
    const unsubscribe = onCacheStatusChange((cached) => {
      setIsOffline(cached);
    });

    return unsubscribe;
  }, []);

  if (!isOffline) {
    return null;
  }

  return (
    <div className="bg-warning/20 border-b border-warning">
      <div className="max-w-7xl mx-auto px-4 py-3">
        <div className="flex items-center gap-3">
          <WifiOff className="w-5 h-5 text-warning flex-shrink-0" />
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-warning" />
              <span className="font-semibold text-warning">Offline Mode</span>
            </div>
            <p className="text-sm text-text-secondary mt-1">
              Unable to connect to Proxmox. Displaying cached data. Some features may be unavailable.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
