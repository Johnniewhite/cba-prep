'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import api from '@/lib/api';
import wsClient from '@/lib/websocket';

interface Message {
  id: string;
  channel_id: string;
  sender_id: string;
  content: string;
  type: string;
  created_at: string;
  updated_at: string;
  sender?: {
    username: string;
    first_name: string;
    last_name: string;
  };
}

export default function ChatPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const channelId = params.channelId as string;
  
  const [messages, setMessages] = useState<Message[]>([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }

    if (channelId) {
      loadMessages();
      setupWebSocket();
      wsClient.joinRoom(`channel:${channelId}`);
    }

    return () => {
      wsClient.leaveRoom(`channel:${channelId}`);
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, [channelId, user]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadMessages = async () => {
    try {
      const messagesData = await api.getMessages(channelId);
      // Ensure messagesData is always an array
      setMessages(Array.isArray(messagesData) ? messagesData : []);
    } catch (error) {
      console.error('Failed to load messages:', error);
      // Set empty array on error
      setMessages([]);
    } finally {
      setLoading(false);
    }
  };

  const setupWebSocket = () => {
    wsClient.on('chat', (message) => {
      if (message.data?.channel_id === channelId) {
        // Add real-time messages from other users
        setMessages(prev => [...prev, message.data]);
      }
    });

    // Typing indicator disabled for now
    // wsClient.on('typing', (message) => {
    //   // Handle typing indicators when properly implemented
    // });
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const sendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newMessage.trim() || sending) return;

    setSending(true);
    
    try {
      const messageData = await api.sendMessage(channelId, newMessage.trim());
      setNewMessage('');
      
      // Add the message to the local state immediately for better UX
      const newMsg: Message = {
        id: messageData.id,
        channel_id: channelId,
        sender_id: messageData.sender_id,
        content: messageData.content,
        type: messageData.type,
        created_at: messageData.created_at,
        updated_at: messageData.updated_at,
        sender: messageData.sender
      };
      setMessages(prev => [...prev, newMsg]);
      
      // Stop typing indicator
      handleTyping(false);
    } catch (error) {
      console.error('Failed to send message:', error);
    } finally {
      setSending(false);
    }
  };

  const handleTyping = (typing: boolean) => {
    // Disable typing indicator for now since WebSocket implementation needs more work
    // if (typing !== isTyping) {
    //   setIsTyping(typing);
    //   wsClient.sendTypingIndicator(channelId, typing);
    // }

    // if (typing) {
    //   // Clear existing timeout
    //   if (typingTimeoutRef.current) {
    //     clearTimeout(typingTimeoutRef.current);
    //   }
      
    //   // Set timeout to stop typing after 1 second of inactivity
    //   typingTimeoutRef.current = setTimeout(() => {
    //     setIsTyping(false);
    //     wsClient.sendTypingIndicator(channelId, false);
    //   }, 1000);
    // }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading chat...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex flex-col">
      {/* Header */}
      <div className="bg-white shadow-sm border-b p-4">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => router.push('/dashboard')}
              className="text-gray-500 hover:text-gray-700"
            >
              ← Back to Dashboard
            </button>
            <h1 className="text-xl font-semibold">Channel Chat</h1>
          </div>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600">
              {user?.username}
            </span>
          </div>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 max-w-4xl mx-auto w-full p-4">
        <div className="bg-white rounded-lg shadow h-full flex flex-col">
          <div className="flex-1 overflow-y-auto p-4 space-y-4">
            {messages.map((message) => (
              <div key={message.id} className="flex space-x-3">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                    {message.sender?.username?.[0]?.toUpperCase() || 'U'}
                  </div>
                </div>
                <div className="flex-1">
                  <div className="flex items-center space-x-2">
                    <span className="font-medium text-sm">
                      {message.sender?.username || `User ${message.sender_id?.slice(0, 8) || 'Unknown'}`}
                    </span>
                    <span className="text-xs text-gray-500">
                      {formatTimestamp(message.created_at)}
                    </span>
                  </div>
                  <div className="mt-1 text-sm text-gray-900">
                    {message.content}
                  </div>
                </div>
              </div>
            ))}
            
            {/* Typing indicator disabled for now
            {typingUsers.length > 0 && (
              <div className="flex space-x-3">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center text-gray-600 text-sm">
                    ...
                  </div>
                </div>
                <div className="flex-1">
                  <div className="text-sm text-gray-500 italic">
                    {typingUsers.length === 1
                      ? `${typingUsers[0]} is typing...`
                      : typingUsers.length === 2
                      ? `${typingUsers.join(' and ')} are typing...`
                      : `${typingUsers.slice(0, -1).join(', ')} and ${typingUsers[typingUsers.length - 1]} are typing...`}
                  </div>
                </div>
              </div>
            )}
            */}
            
            <div ref={messagesEndRef} />
          </div>

          {/* Message Input */}
          <div className="border-t p-4">
            <form onSubmit={sendMessage} className="flex space-x-4">
              <input
                type="text"
                value={newMessage}
                onChange={(e) => setNewMessage(e.target.value)}
                placeholder="Type a message..."
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                disabled={sending}
              />
              <button
                type="submit"
                disabled={!newMessage.trim() || sending}
                className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {sending ? 'Sending...' : 'Send'}
              </button>
            </form>
          </div>
        </div>
      </div>

      {messages.length === 0 && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center text-gray-500">
            <p className="text-lg mb-2">No messages yet</p>
            <p className="text-sm">Be the first to start the conversation!</p>
          </div>
        </div>
      )}
    </div>
  );
}