import type { Metadata } from 'next';
import './globals.css';
import TopBar from '@/components/layout/TopBar';
import Sidebar from '@/components/layout/Sidebar';
import OfflineBanner from '@/components/layout/OfflineBanner';
import ErrorBoundary from '@/components/ErrorBoundary';

export const metadata: Metadata = {
  title: 'ProxiCloud - Proxmox Management Console',
  description: 'AWS-style management console for Proxmox VE',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="antialiased">
        <ErrorBoundary>
          <div className="flex flex-col h-screen">
            <TopBar />
            <OfflineBanner />
            <div className="flex flex-1 overflow-hidden">
              <Sidebar />
              <main className="flex-1 overflow-y-auto p-6 bg-background">
                {children}
              </main>
            </div>
          </div>
        </ErrorBoundary>
      </body>
    </html>
  );
}
