import React from 'react';
import { UserIcon } from '@heroicons/react/24/outline';

interface Message {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  senderId?: string;
}

interface ChatMessageProps {
  message: Message;
}

export default function ChatMessage({ message }: ChatMessageProps) {
  const isMe = message.role === 'user';

  return (
    <div className={`mb-6 animate-fade-in ${isMe ? 'flex justify-end' : ''}`}>
      <div className={`flex gap-3 max-w-[80%] ${isMe ? 'flex-row-reverse' : ''}`}>
        {/* Avatar */}
        <div className={`flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center ${
          isMe 
            ? 'bg-[var(--accent-primary)]' 
            : 'bg-[var(--bg-tertiary)] border border-[var(--border-primary)]'
        }`}>
          {isMe ? (
            <UserIcon className="w-5 h-5 text-white" />
          ) : (
            <div className="w-5 h-5 flex items-center justify-center text-[var(--text-secondary)] font-bold text-xs">
              {message.senderId ? message.senderId.slice(0, 2).toUpperCase() : 'Rx'}
            </div>
          )}
        </div>

        {/* Message Content */}
        <div className="flex-1 min-w-0">
          <div className={`rounded-lg px-4 py-3 shadow-sm ${
            isMe
              ? 'bg-[var(--accent-primary)] text-white'
              : 'bg-[var(--bg-secondary)] border border-[var(--border-primary)] text-[var(--text-primary)]'
          }`}>
            <div className="text-sm leading-relaxed whitespace-pre-wrap break-words">
              {message.content}
            </div>
          </div>
          
          <div className={`text-xs text-[var(--text-tertiary)] mt-1 ${isMe ? 'text-right' : 'text-left'}`}>
            {new Date(message.timestamp).toLocaleTimeString([], { 
                hour: '2-digit', 
                minute: '2-digit' 
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
