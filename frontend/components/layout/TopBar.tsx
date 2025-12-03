'use client';

import { Server, Bell } from 'lucide-react';

export default function TopBar() {
  return (
    <header className="h-16 bg-surface border-b border-border flex items-center justify-between px-6 sticky top-0 z-10">
      <div className="flex items-center gap-3">
        <Server className="w-6 h-6 text-primary" />
        <h1 className="text-xl font-bold text-text-primary">ProxiCloud</h1>
      </div>
      
      <div className="flex items-center gap-4">
        <button className="p-2 hover:bg-surface-elevated rounded-lg transition-colors relative">
          <Bell className="w-5 h-5 text-text-secondary" />
        </button>
        
        <div className="flex items-center gap-3 pl-4 border-l border-border">
          <div className="text-right">
            <div className="text-sm font-medium text-text-primary">Admin</div>
            <div className="text-xs text-text-muted">Connected</div>
          </div>
          <div className="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-white text-sm font-medium">
            A
          </div>
        </div>
      </div>
    </header>
  );
}
