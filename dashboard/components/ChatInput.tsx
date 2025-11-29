import React, { useState } from 'react';
import { PaperAirplaneIcon } from '@heroicons/react/24/solid';

interface ChatInputProps {
  onSubmit: (message: string) => void;
  disabled?: boolean;
}

export default function ChatInput({ onSubmit, disabled }: ChatInputProps) {
  const [input, setInput] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim() && !disabled) {
      onSubmit(input.trim());
      setInput('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="relative">
      <textarea
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        placeholder="Send a message..."
        rows={1}
        className="w-full resize-none bg-[var(--bg-tertiary)] text-[var(--text-primary)] border border-[var(--border-primary)] rounded-xl px-4 py-3 pr-12 focus:outline-none focus:border-[var(--accent-primary)] placeholder-[var(--text-tertiary)] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        style={{
          minHeight: '52px',
          maxHeight: '200px',
        }}
      />
      <button
        type="submit"
        disabled={!input.trim() || disabled}
        className="absolute right-2 bottom-2 p-2 bg-[var(--accent-primary)] hover:bg-[#4897E8] text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:bg-[var(--accent-primary)] transition-all"
      >
        <PaperAirplaneIcon className="w-5 h-5" />
      </button>
    </form>
  );
}
