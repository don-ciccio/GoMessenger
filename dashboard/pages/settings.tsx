import React, { useState } from 'react';
import Layout from '../components/Layout';
import { 
  Cog6ToothIcon,
  BellIcon,
  ShieldCheckIcon,
  KeyIcon,
  ServerIcon,
  CheckCircleIcon
} from '@heroicons/react/24/outline';

export default function Settings() {
  const [notifications, setNotifications] = useState(true);
  const [autoSave, setAutoSave] = useState(true);
  const [saved, setSaved] = useState(false);

  const handleSave = () => {
    // Simulate save
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  return (
    <Layout title="Settings">
      <div className="p-6 space-y-6 max-w-4xl">
        {/* General Settings */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <Cog6ToothIcon className="w-6 h-6 text-[var(--accent-primary)]" />
            <h3 className="text-lg font-semibold">General Settings</h3>
          </div>
          
          <div className="space-y-4">
            <div className="flex items-center justify-between py-3 border-b border-[var(--border-primary)]">
              <div>
                <div className="font-medium">Notifications</div>
                <div className="text-sm text-[var(--text-secondary)]">Receive system notifications</div>
              </div>
              <button
                onClick={() => setNotifications(!notifications)}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  notifications ? 'bg-[var(--accent-primary)]' : 'bg-[var(--bg-tertiary)]'
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    notifications ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>

            <div className="flex items-center justify-between py-3 border-b border-[var(--border-primary)]">
              <div>
                <div className="font-medium">Auto-save</div>
                <div className="text-sm text-[var(--text-secondary)]">Automatically save changes</div>
              </div>
              <button
                onClick={() => setAutoSave(!autoSave)}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                  autoSave ? 'bg-[var(--accent-primary)]' : 'bg-[var(--bg-tertiary)]'
                }`}
              >
                <span
                  className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                    autoSave ? 'translate-x-6' : 'translate-x-1'
                  }`}
                />
              </button>
            </div>
          </div>
        </div>

        {/* API Configuration */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <KeyIcon className="w-6 h-6 text-[var(--accent-primary)]" />
            <h3 className="text-lg font-semibold">API Configuration</h3>
          </div>
          
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-2">Backend API URL</label>
              <input
                type="text"
                value="https://ai-support-backend-z4eq.onrender.com"
                readOnly
                className="input bg-[var(--bg-tertiary)] text-[var(--text-secondary)]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-2">RAG Service URL</label>
              <input
                type="text"
                value="https://ai-support-rag.onrender.com"
                readOnly
                className="input bg-[var(--bg-tertiary)] text-[var(--text-secondary)]"
              />
            </div>
          </div>
        </div>

        {/* System Information */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <ServerIcon className="w-6 h-6 text-[var(--accent-primary)]" />
            <h3 className="text-lg font-semibold">System Information</h3>
          </div>
          
          <div className="space-y-3">
            <div className="flex justify-between py-2 border-b border-[var(--border-primary)]">
              <span className="text-[var(--text-secondary)]">Version</span>
              <span className="font-medium">1.0.0</span>
            </div>
            <div className="flex justify-between py-2 border-b border-[var(--border-primary)]">
              <span className="text-[var(--text-secondary)]">Platform</span>
              <span className="font-medium">Render.com</span>
            </div>
            <div className="flex justify-between py-2 border-b border-[var(--border-primary)]">
              <span className="text-[var(--text-secondary)]">LLM Provider</span>
              <span className="font-medium">Groq (llama-3.1-70b)</span>
            </div>
            <div className="flex justify-between py-2 border-b border-[var(--border-primary)]">
              <span className="text-[var(--text-secondary)]">Embeddings</span>
              <span className="font-medium">OpenAI (text-embedding-3-small)</span>
            </div>
            <div className="flex justify-between py-2">
              <span className="text-[var(--text-secondary)]">Vector Database</span>
              <span className="font-medium">Qdrant Cloud</span>
            </div>
          </div>
        </div>

        {/* Security */}
        <div className="card">
          <div className="flex items-center gap-3 mb-4">
            <ShieldCheckIcon className="w-6 h-6 text-[var(--accent-primary)]" />
            <h3 className="text-lg font-semibold">Security</h3>
          </div>
          
          <div className="space-y-3">
            <div className="flex items-center gap-2 p-3 bg-green-500/10 border border-green-500/20 rounded-lg">
              <CheckCircleIcon className="w-5 h-5 text-green-500" />
              <span className="text-sm text-green-500">All services are secure and encrypted</span>
            </div>
            <button className="btn-secondary w-full">
              Change API Keys
            </button>
            <button className="btn-secondary w-full">
              View Activity Logs
            </button>
          </div>
        </div>

        {/* Save Button */}
        <div className="flex items-center gap-3">
          <button
            onClick={handleSave}
            className="btn-primary px-6"
          >
            Save Changes
          </button>
          {saved && (
            <div className="flex items-center gap-2 text-[var(--accent-success)] animate-fade-in">
              <CheckCircleIcon className="w-5 h-5" />
              <span className="text-sm">Settings saved successfully</span>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}
