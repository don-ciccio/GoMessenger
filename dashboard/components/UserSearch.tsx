import React, { useState } from 'react';
import { MagnifyingGlassIcon, XMarkIcon, UserPlusIcon } from '@heroicons/react/24/outline';
import { UserCircleIcon } from '@heroicons/react/24/solid';

interface User {
  id: string;
  username: string;
}

interface UserSearchProps {
  currentUserId: string;
  onStartConversation: (userId: string, username: string) => void;
}

export default function UserSearch({ currentUserId, onStartConversation }: UserSearchProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<User[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const handleSearch = async (query: string) => {
    setSearchQuery(query);
    
    if (query.trim().length < 2) {
      setSearchResults([]);
      return;
    }

    setIsSearching(true);
    try {
      const response = await fetch(`http://localhost:8080/users/search?q=${encodeURIComponent(query)}`);
      if (response.ok) {
        const data = await response.json();
        // Filter out current user
        const filtered = (data || []).filter((user: User) => user.id !== currentUserId);
        setSearchResults(filtered);
      }
    } catch (error) {
      console.error('Failed to search users:', error);
    } finally {
      setIsSearching(false);
    }
  };

  const handleSelectUser = async (user: User) => {
    // Create or get conversation with this user
    try {
      const response = await fetch('http://localhost:8080/conversations', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          participants: [currentUserId, user.id]
        })
      });
      
      if (response.ok) {
        const conversation = await response.json();
        onStartConversation(user.id, user.username);
        setIsOpen(false);
        setSearchQuery('');
        setSearchResults([]);
      }
    } catch (error) {
      console.error('Failed to create conversation:', error);
    }
  };

  return (
    <>
      {/* Search Button */}
      <button
        onClick={() => setIsOpen(true)}
        className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-[var(--bg-tertiary)] text-[var(--text-secondary)] transition-colors border border-[var(--border-primary)]"
      >
        <UserPlusIcon className="w-5 h-5" />
        <span>New Chat</span>
      </button>

      {/* Search Modal */}
      {isOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={() => setIsOpen(false)}>
          <div 
            className="bg-[var(--bg-secondary)] rounded-lg shadow-xl w-full max-w-md mx-4"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Header */}
            <div className="flex items-center justify-between p-4 border-b border-[var(--border-primary)]">
              <h2 className="text-lg font-semibold">Search Users</h2>
              <button
                onClick={() => setIsOpen(false)}
                className="p-1 hover:bg-[var(--bg-tertiary)] rounded-lg transition-colors"
              >
                <XMarkIcon className="w-5 h-5" />
              </button>
            </div>

            {/* Search Input */}
            <div className="p-4 border-b border-[var(--border-primary)]">
              <div className="relative">
                <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-[var(--text-tertiary)]" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => handleSearch(e.target.value)}
                  placeholder="Search by username..."
                  autoFocus
                  className="w-full pl-10 pr-4 py-2 bg-[var(--bg-tertiary)] border border-[var(--border-primary)] rounded-lg text-sm focus:outline-none focus:border-[var(--accent-primary)]"
                />
              </div>
            </div>

            {/* Results */}
            <div className="max-h-96 overflow-y-auto">
              {isSearching ? (
                <div className="p-8 text-center text-[var(--text-tertiary)]">
                  Searching...
                </div>
              ) : searchQuery.length < 2 ? (
                <div className="p-8 text-center text-[var(--text-tertiary)] text-sm">
                  Type at least 2 characters to search
                </div>
              ) : searchResults.length === 0 ? (
                <div className="p-8 text-center text-[var(--text-tertiary)] text-sm">
                  No users found
                </div>
              ) : (
                <div className="divide-y divide-[var(--border-primary)]">
                  {searchResults.map((user) => (
                    <button
                      key={user.id}
                      onClick={() => handleSelectUser(user)}
                      className="w-full p-4 hover:bg-[var(--bg-tertiary)] transition-colors flex items-center gap-3 text-left"
                    >
                      <UserCircleIcon className="w-10 h-10 text-[var(--text-tertiary)] flex-shrink-0" />
                      <div className="flex-1 min-w-0">
                        <div className="font-medium text-[var(--text-primary)] truncate">
                          {user.username}
                        </div>
                        <div className="text-xs text-[var(--text-tertiary)] truncate">
                          ID: {user.id.slice(0, 12)}
                        </div>
                      </div>
                      <UserPlusIcon className="w-5 h-5 text-[var(--text-secondary)] flex-shrink-0" />
                    </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
