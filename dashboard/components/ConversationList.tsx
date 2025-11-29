import React from 'react';
import { UserCircleIcon } from '@heroicons/react/24/solid';

interface Conversation {
  id: string;
  participants: string[];
  last_message: string;
  last_message_at: string;
}

interface ConversationListProps {
  conversations: Conversation[];
  currentUserId: string;
  activeConversationId?: string;
  onSelectConversation: (conversationId: string) => void;
  usernames: Record<string, string>;
  onlineUsers: Set<string>;
}

export default function ConversationList({
  conversations,
  currentUserId,
  activeConversationId,
  onSelectConversation,
  usernames,
  onlineUsers,
}: ConversationListProps) {
  const getOtherUser = (participants: string[]) => {
    return participants.find(id => id !== currentUserId) || 'Unknown';
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="p-3 border-b border-[var(--border-primary)]">
        <h2 className="font-semibold text-sm text-[var(--text-secondary)]">Conversations</h2>
      </div>
      
      {conversations.length === 0 ? (
        <div className="p-4 text-center text-[var(--text-tertiary)] text-sm">
          No conversations yet
        </div>
      ) : (
        <div className="divide-y divide-[var(--border-primary)]">
          {conversations.map((conv) => {
            const isActive = conv.id === activeConversationId;
            const otherUserId = getOtherUser(conv.participants);
            const username = usernames[otherUserId] || `User ${otherUserId.slice(0, 8)}`;
            const isOnline = onlineUsers.has(otherUserId);
            
            return (
              <button
                key={conv.id}
                onClick={() => onSelectConversation(conv.id)}
                className={`w-full text-left p-3 hover:bg-[var(--bg-tertiary)] transition-colors ${
                  isActive ? 'bg-[var(--bg-tertiary)] border-l-2 border-blue-500' : ''
                }`}
              >
                <div className="flex items-start gap-3">
                  <div className="relative">
                    <UserCircleIcon className="w-10 h-10 text-[var(--text-tertiary)] flex-shrink-0" />
                    <span className={`absolute bottom-0 right-0 w-3 h-3 border-2 border-[var(--bg-primary)] rounded-full ${isOnline ? 'bg-green-500' : 'bg-red-500'}`}></span>
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-baseline justify-between gap-2 mb-1">
                      <span className="font-medium text-sm text-[var(--text-primary)] truncate">
                        {username}
                      </span>
                      {conv.last_message_at && (
                        <span className="text-xs text-[var(--text-tertiary)] flex-shrink-0">
                          {formatTime(conv.last_message_at)}
                        </span>
                      )}
                    </div>
                    
                    {conv.last_message && (
                      <p className="text-sm text-[var(--text-secondary)] truncate">
                        {conv.last_message}
                      </p>
                    )}
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
