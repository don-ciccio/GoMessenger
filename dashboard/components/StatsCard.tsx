import React from 'react';

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
}

export default function StatsCard({ title, value, icon, change, changeType = 'neutral' }: StatsCardProps) {
  return (
    <div className="card">
      <div className="flex items-start justify-between">
        <div>
          <div className="text-sm text-[var(--text-secondary)] mb-1">{title}</div>
          <div className="text-2xl font-semibold text-[var(--text-primary)]">{value}</div>
          {change && (
            <div className={`text-sm mt-1 ${
              changeType === 'positive' ? 'text-[var(--accent-success)]' :
              changeType === 'negative' ? 'text-[var(--accent-danger)]' :
              'text-[var(--text-tertiary)]'
            }`}>
              {change}
            </div>
          )}
        </div>
        <div className="p-2 bg-[var(--bg-tertiary)] rounded-lg text-[var(--accent-primary)]">
          {icon}
        </div>
      </div>
    </div>
  );
}
