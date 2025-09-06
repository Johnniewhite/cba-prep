import { EventEmitter } from 'events';

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/api/v1/ws';

export type MessageType = 'chat' | 'task_update' | 'user_status' | 'notification' | 'typing' | 'presence';

export interface WSMessage {
  type: MessageType;
  room?: string;
  user_id?: string;
  data?: any;
  timestamp: string;
}

class WebSocketClient extends EventEmitter {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private token: string | null = null;

  connect(token: string) {
    this.token = token;
    this.initConnection();
  }

  private initConnection() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return;
    }

    const wsUrl = `${WS_URL}${this.token ? `?token=${this.token}` : ''}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket connected');
      this.reconnectAttempts = 0;
      this.emit('connected');
      this.startHeartbeat();
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.emit('error', error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket disconnected');
      this.stopHeartbeat();
      this.emit('disconnected');
      this.attemptReconnect();
    };
  }

  private handleMessage(message: WSMessage) {
    this.emit('message', message);
    
    switch (message.type) {
      case 'chat':
        this.emit('chat', message);
        break;
      case 'task_update':
        this.emit('task_update', message);
        break;
      case 'typing':
        this.emit('typing', message);
        break;
      case 'presence':
        this.emit('presence', message);
        break;
      case 'notification':
        this.emit('notification', message);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }

  sendMessage(type: MessageType, data: any, room?: string) {
    if (this.ws?.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return;
    }

    const message: WSMessage = {
      type,
      data,
      room,
      timestamp: new Date().toISOString(),
    };

    this.ws.send(JSON.stringify(message));
  }

  sendChatMessage(content: string, channelId: string) {
    this.sendMessage('chat', { content, channel_id: channelId }, `channel:${channelId}`);
  }

  sendTypingIndicator(channelId: string, isTyping: boolean) {
    this.sendMessage('typing', { is_typing: isTyping, channel_id: channelId }, `channel:${channelId}`);
  }

  joinRoom(room: string) {
    this.sendMessage('notification', { action: 'join_room', room }, room);
  }

  leaveRoom(room: string) {
    this.sendMessage('notification', { action: 'leave_room', room }, room);
  }

  private attemptReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      this.emit('max_reconnect_failed');
      return;
    }

    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    console.log(`Attempting to reconnect in ${delay}ms...`);
    
    setTimeout(() => {
      this.initConnection();
    }, delay);
  }

  private startHeartbeat() {
    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }

  disconnect() {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const wsClient = new WebSocketClient();
export default wsClient;