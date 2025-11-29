import React, { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/router';
import ChatMessage from '../components/ChatMessage';
import ChatInput from '../components/ChatInput';
import ConversationList from '../components/ConversationList';
import UserSearch from '../components/UserSearch';
import { 
  Bars3Icon,
  ChatBubbleLeftIcon,
  Cog6ToothIcon,
  UserIcon,
  ArrowRightOnRectangleIcon,
  MagnifyingGlassIcon
} from '@heroicons/react/24/outline';

interface Message {
  id: string;
  role: 'user' | 'assistant'; // 'user' = me, 'assistant' = them
  content: string;
  timestamp: Date;
  senderId?: string;
}

interface Conversation {
  id: string;
  participants: string[];
  last_message: string;
  last_message_at: string;
}

export default function Home() {
  const router = useRouter();
  const [messages, setMessages] = useState<Message[]>([]);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeConversationId, setActiveConversationId] = useState<string>('');
  const [usernames, setUsernames] = useState<Record<string, string>>({});
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [targetUserId, setTargetUserId] = useState('');
  const [myUserId, setMyUserId] = useState('');
  const [connected, setConnected] = useState(false);
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set());
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Auth Check
    const token = localStorage.getItem('token');
    const userId = localStorage.getItem('userId');
    
    if (!token || !userId) {
      router.push('/login');
      return;
    }

    setMyUserId(userId);
    
    // Fetch conversations
    fetchConversations(userId);

    // WebSocket Connection
    const backendUrl = process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';
    const wsUrl = backendUrl.replace(/^http/, 'ws') + '/ws';
    const ws = new WebSocket(`${wsUrl}?token=${token}`);
    
    ws.onopen = () => {
      console.log('Connected to WebSocket');
      setConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        
        // Handle presence events
        if (data.type === 'online_users') {
          setOnlineUsers(new Set(data.user_ids));
        } else if (data.type === 'user_status') {
          setOnlineUsers(prev => {
            const next = new Set(prev);
            if (data.online) {
              next.add(data.user_id);
            } else {
              next.delete(data.user_id);
            }
            return next;
          });
        }
        // Handle incoming messages
        else if (data.content) {
            const isMe = data.sender_id === userId;
            const newMessage: Message = {
                id: data.id || Date.now().toString(),
                role: isMe ? 'user' : 'assistant',
                content: data.content,
                timestamp: new Date(data.timestamp ? data.timestamp * 1000 : Date.now()),
                senderId: data.sender_id
            };
            setMessages((prev) => [...prev, newMessage]);
            
            // Refresh conversations to update last message
            if (data.conversation_id) {
              fetchConversations(userId);
            }
        }
      } catch (e) {
        console.error('Error parsing message:', e);
      }
    };

    ws.onclose = () => {
      console.log('Disconnected');
      setConnected(false);
    };

    wsRef.current = ws;

    return () => {
      ws.close();
    };
  }, [router]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const fetchUsernames = async (userIds: string[]) => {
    const missingIds = userIds.filter(id => !usernames[id] && id !== myUserId);
    if (missingIds.length === 0) return;

    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/users/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids: missingIds })
      });
      if (response.ok) {
        const users = await response.json();
        setUsernames(prev => {
          const next = { ...prev };
          users.forEach((u: any) => {
            next[u.id] = u.username;
          });
          return next;
        });
      }
    } catch (error) {
      console.error('Failed to fetch usernames:', error);
    }
  };

  const fetchConversations = async (userId: string) => {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/conversations?user_id=${userId}`);
      if (response.ok) {
        const data = await response.json();
        setConversations(data || []);
        
        // Fetch usernames for all participants
        const allParticipantIds = new Set<string>();
        (data || []).forEach((c: Conversation) => {
            c.participants.forEach(id => allParticipantIds.add(id));
        });
        fetchUsernames(Array.from(allParticipantIds));
      }
    } catch (error) {
      console.error('Failed to fetch conversations:', error);
    }
  };

  const handleSelectConversation = async (conversationId: string) => {
    setActiveConversationId(conversationId);
    setMessages([]); // Clear current messages
    
    // Find the other participant
    const conversation = conversations.find(c => c.id === conversationId);
    if (conversation) {
      const otherUserId = conversation.participants.find(id => id !== myUserId);
      setTargetUserId(otherUserId || '');
    }
    
    // Load message history
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_BACKEND_URL}/conversations/${conversationId}/messages?limit=50`);
      if (response.ok) {
        const data = await response.json();
        const loadedMessages: Message[] = (data || []).map((msg: any) => ({
          id: msg.id,
          role: msg.sender_id === myUserId ? 'user' : 'assistant',
          content: msg.content,
          timestamp: new Date(msg.timestamp * 1000),
          senderId: msg.sender_id
        }));
        setMessages(loadedMessages);
      }
    } catch (error) {
      console.error('Failed to load messages:', error);
    }
  };

  const handleStartConversation = (userId: string, username: string) => {
    // Store username
    setUsernames(prev => ({ ...prev, [userId]: username }));
    
    // Refresh conversations to get the new one
    fetchConversations(myUserId);
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSubmitMessage = (content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      alert('Not connected to chat server');
      return;
    }
    
    // Check if we have an active conversation or target user
    if (!targetUserId && !activeConversationId) {
      alert('Please select a conversation or start a new chat');
      return;
    }

    const payload: any = {
      sender_id: myUserId,
      content: content
    };
    
    // Add conversation_id if we have an active conversation
    if (activeConversationId) {
      payload.conversation_id = activeConversationId;
    } else if (targetUserId) {
      // Fallback for direct user messaging (will auto-create conversation)
      payload.receiver_id = targetUserId;
    }

    const message = {
      type: 'chat_message',
      sender_id: myUserId,
      payload: payload
    };

    wsRef.current.send(JSON.stringify(message));
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('userId');
    localStorage.removeItem('username');
    router.push('/login');
  };

  return (
    <div className="flex h-screen bg-[var(--bg-primary)] text-[var(--text-primary)]">
      {/* Sidebar */}
      <div
        className={`${
          sidebarOpen ? 'w-64' : 'w-0'
        } transition-all duration-300 bg-[var(--bg-secondary)] border-r border-[var(--border-primary)] flex flex-col overflow-hidden`}
      >
        <div className="p-4 border-b border-[var(--border-primary)]">
            <div className="font-bold text-lg mb-1">GoMessenger</div>
            <div className="text-xs text-[var(--text-tertiary)] truncate">ID: {myUserId.slice(0, 8)}</div>
            <div className={`text-xs mt-2 flex items-center gap-2 ${connected ? 'text-green-500' : 'text-red-500'}`}>
                <span className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`}></span>
                {connected ? 'Online' : 'Offline'}
            </div>
        </div>

        <div className="p-3 border-b border-[var(--border-primary)]">
          <UserSearch 
            currentUserId={myUserId}
            onStartConversation={handleStartConversation}
          />
        </div>

        <ConversationList
          conversations={conversations}
          currentUserId={myUserId}
          activeConversationId={activeConversationId}
          onSelectConversation={handleSelectConversation}
          usernames={usernames}
          onlineUsers={onlineUsers}
        />

        <div className="border-t border-[var(--border-primary)] p-3">
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-[var(--bg-tertiary)] text-[var(--text-secondary)] transition-colors"
          >
            <ArrowRightOnRectangleIcon className="w-5 h-5" />
            <span>Logout</span>
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Header */}
        <header className="h-14 border-b border-[var(--border-primary)] flex items-center px-4 gap-3">
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="p-2 hover:bg-[var(--bg-hover)] rounded-lg transition-colors"
          >
            <Bars3Icon className="w-5 h-5" />
          </button>
          <div className="text-sm font-medium">
            {targetUserId ? `Chatting with ${usernames[targetUserId] || targetUserId.slice(0, 8)}` : 'Select a user to chat'}
          </div>
        </header>

        {/* Messages */}
        <div className="flex-1 overflow-y-auto bg-[var(--bg-primary)]">
          <div className="max-w-3xl mx-auto px-4 py-6">
            {messages.map((message) => (
              <ChatMessage
                key={message.id}
                message={message}
              />
            ))}
            <div ref={messagesEndRef} />
          </div>
        </div>

        {/* Input */}
        <div className="border-t border-[var(--border-primary)] p-4 bg-[var(--bg-secondary)]">
          <div className="max-w-3xl mx-auto">
            <ChatInput
              onSubmit={handleSubmitMessage}
              disabled={!connected || !targetUserId}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
