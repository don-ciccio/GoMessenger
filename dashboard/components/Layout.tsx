import React, { useState } from 'react';
import { useRouter } from 'next/router';
import {
  Bars3Icon,
  ChatBubbleLeftIcon,
  ChartBarIcon,
  DocumentTextIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline';

interface LayoutProps {
  children: React.ReactNode;
  title: string;
}

export default function Layout({ children, title }: LayoutProps) {
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const navigation = [
    { name: 'Chat', href: '/', icon: ChatBubbleLeftIcon },
    { name: 'Analytics', href: '/analytics', icon: ChartBarIcon },
    { name: 'Documents', href: '/documents', icon: DocumentTextIcon },
    { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
  ];

  return (
    <div className="flex h-screen bg-[var(--bg-primary)] text-[var(--text-primary)]">
      {/* Sidebar */}
      <div
        className={`${
          sidebarOpen ? 'w-64' : 'w-0'
        } transition-all duration-300 bg-[var(--bg-secondary)] border-r border-[var(--border-primary)] flex flex-col overflow-hidden`}
      >
        <div className="p-4 border-b border-[var(--border-primary)]">
          <div className="font-semibold text-lg">AI Support</div>
          <div className="text-xs text-[var(--text-tertiary)] mt-1">Admin Dashboard</div>
        </div>

        <nav className="flex-1 p-3 space-y-1 overflow-y-auto">
          {navigation.map((item) => {
            const isActive = router.pathname === item.href;
            return (
              <button
                key={item.name}
                onClick={() => router.push(item.href)}
                className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${
                  isActive
                    ? 'bg-[var(--bg-hover)] text-[var(--text-primary)]'
                    : 'text-[var(--text-secondary)] hover:bg-[var(--bg-hover)] hover:text-[var(--text-primary)]'
                }`}
              >
                <item.icon className="w-5 h-5" />
                <span>{item.name}</span>
              </button>
            );
          })}
        </nav>

        <div className="p-4 border-t border-[var(--border-primary)]">
          <div className="text-xs text-[var(--text-tertiary)]">
            <div className="font-semibold text-[var(--text-primary)] mb-1">AI Support Assistant</div>
            <div>Powered by Groq & Next.js</div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-14 border-b border-[var(--border-primary)] flex items-center px-4 gap-3 flex-shrink-0">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
          <div className="text-sm font-medium">{title}</div>
        </header>

        {/* Content */}
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </div>
    </div>
  );
}
