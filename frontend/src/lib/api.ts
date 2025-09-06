const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

class ApiClient {
  private token: string | null = null;

  constructor() {
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('access_token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
    }
  }

  async request<T = any>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_URL}${endpoint}`;
    
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (response.status === 401) {
      this.clearToken();
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Auth endpoints
  async register(data: {
    email: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
  }) {
    const response = await this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
      if (response.refresh_token && typeof window !== 'undefined') {
        localStorage.setItem('refresh_token', response.refresh_token);
      }
    }
    
    return response;
  }

  async login(email_or_username: string, password: string) {
    const response = await this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email_or_username, password }),
    });
    
    if (response.access_token) {
      this.setToken(response.access_token);
      if (response.refresh_token && typeof window !== 'undefined') {
        localStorage.setItem('refresh_token', response.refresh_token);
      }
    }
    
    return response;
  }

  async logout() {
    await this.request('/auth/logout', { method: 'POST' });
    this.clearToken();
  }

  // User endpoints
  async getCurrentUser() {
    return this.request('/users/me');
  }

  async updateCurrentUser(data: any) {
    return this.request('/users/me', {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  // Team endpoints
  async getTeams() {
    return this.request('/teams');
  }

  async createTeam(data: { name: string; description?: string }) {
    return this.request('/teams', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getTeam(teamId: string) {
    return this.request(`/teams/${teamId}`);
  }

  async getTeamMembers(teamId: string) {
    return this.request(`/teams/${teamId}/members`);
  }

  async inviteTeamMember(teamId: string, email: string, role: string) {
    return this.request(`/teams/${teamId}/members`, {
      method: 'POST',
      body: JSON.stringify({ email, role }),
    });
  }

  // Channel endpoints
  async getChannels(teamId: string) {
    return this.request(`/teams/${teamId}/channels`);
  }

  async createChannel(teamId: string, data: {
    name: string;
    description?: string;
    type: string;
    is_private?: boolean;
  }) {
    return this.request(`/teams/${teamId}/channels`, {
      method: 'POST',
      body: JSON.stringify({ ...data, team_id: teamId }),
    });
  }

  // Message endpoints
  async getMessages(channelId: string, limit = 50) {
    return this.request(`/channels/${channelId}/messages?limit=${limit}`);
  }

  async sendMessage(channelId: string, content: string, type = 'text') {
    return this.request(`/channels/${channelId}/messages`, {
      method: 'POST',
      body: JSON.stringify({
        channel_id: channelId,
        content,
        type,
      }),
    });
  }

  async updateMessage(messageId: string, content: string) {
    return this.request(`/messages/${messageId}`, {
      method: 'PUT',
      body: JSON.stringify({ content }),
    });
  }

  async deleteMessage(messageId: string) {
    return this.request(`/messages/${messageId}`, {
      method: 'DELETE',
    });
  }

  // Task endpoints
  async getTasks(teamId: string, filters?: any) {
    const params = new URLSearchParams(filters).toString();
    return this.request(`/teams/${teamId}/tasks${params ? `?${params}` : ''}`);
  }

  async createTask(teamId: string, data: {
    title: string;
    description?: string;
    priority: string;
    assignee_id?: string;
    due_date?: string;
    tags?: string[];
  }) {
    return this.request(`/teams/${teamId}/tasks`, {
      method: 'POST',
      body: JSON.stringify({ ...data, team_id: teamId }),
    });
  }

  async getTask(taskId: string) {
    return this.request(`/tasks/${taskId}`);
  }

  async updateTask(taskId: string, data: any) {
    return this.request(`/tasks/${taskId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteTask(taskId: string) {
    return this.request(`/tasks/${taskId}`, {
      method: 'DELETE',
    });
  }

  async getTaskComments(taskId: string) {
    return this.request(`/tasks/${taskId}/comments`);
  }

  async createTaskComment(taskId: string, content: string) {
    return this.request(`/tasks/${taskId}/comments`, {
      method: 'POST',
      body: JSON.stringify({ task_id: taskId, content }),
    });
  }
}

export const api = new ApiClient();
export default api;