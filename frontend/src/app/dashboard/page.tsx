'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';
import wsClient from '@/lib/websocket';

interface Team {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  created_at: string;
}

interface Channel {
  id: string;
  team_id: string;
  name: string;
  description: string;
  type: string;
  is_private: boolean;
}

interface Task {
  id: string;
  team_id: string;
  title: string;
  description: string;
  status: string;
  priority: string;
  assignee_id?: string;
  created_by: string;
  due_date?: string;
  created_at: string;
}

interface TeamMember {
  user_id: string;
  role: string;
  joined_at: string;
  updated_at: string;
  user: {
    email: string;
    username: string;
    first_name: string;
    last_name: string;
    avatar?: string;
  };
}

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [selectedChannel, setSelectedChannel] = useState<Channel | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'chat' | 'tasks' | 'members'>('chat');
  const [showNewTeamForm, setShowNewTeamForm] = useState(false);
  const [showInviteMemberForm, setShowInviteMemberForm] = useState(false);
  const [newTeamName, setNewTeamName] = useState('');
  const [memberEmail, setMemberEmail] = useState('');
  const [memberRole, setMemberRole] = useState('member');
  const [wsConnected, setWsConnected] = useState(false);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }
    
    if (user) {
      loadTeams();
      setupWebSocket();
    }
  }, [user, authLoading]);

  useEffect(() => {
    if (selectedTeam) {
      loadChannels(selectedTeam.id);
      loadTasks(selectedTeam.id);
      loadTeamMembers(selectedTeam.id);
    }
  }, [selectedTeam]);

  const setupWebSocket = () => {
    wsClient.on('connected', () => {
      console.log('WebSocket connected');
      setWsConnected(true);
    });

    wsClient.on('disconnected', () => {
      console.log('WebSocket disconnected');
      setWsConnected(false);
    });

    wsClient.on('chat', (message) => {
      console.log('New chat message:', message);
      // Handle new chat messages
    });

    wsClient.on('task_update', (message) => {
      console.log('Task update:', message);
      // Reload tasks when there's an update
      if (selectedTeam) {
        loadTasks(selectedTeam.id);
      }
    });
  };

  const loadTeams = async () => {
    try {
      const teamsData = await api.getTeams();
      // Ensure teamsData is always an array
      const teams = Array.isArray(teamsData) ? teamsData : [];
      setTeams(teams);
      if (teams.length > 0 && !selectedTeam) {
        setSelectedTeam(teams[0]);
      }
    } catch (error) {
      console.error('Failed to load teams:', error);
      // Ensure teams is always an empty array on error
      setTeams([]);
    } finally {
      setLoading(false);
    }
  };

  const loadChannels = async (teamId: string) => {
    try {
      const channelsData = await api.getChannels(teamId);
      const channels = Array.isArray(channelsData) ? channelsData : [];
      setChannels(channels);
      if (channels.length > 0 && !selectedChannel) {
        setSelectedChannel(channels[0]);
      }
    } catch (error) {
      console.error('Failed to load channels:', error);
      setChannels([]);
    }
  };

  const loadTasks = async (teamId: string) => {
    try {
      const tasksData = await api.getTasks(teamId);
      const tasks = Array.isArray(tasksData) ? tasksData : [];
      setTasks(tasks);
    } catch (error) {
      console.error('Failed to load tasks:', error);
      setTasks([]);
    }
  };

  const loadTeamMembers = async (teamId: string) => {
    try {
      const membersData = await api.getTeamMembers(teamId);
      const members = Array.isArray(membersData) ? membersData : [];
      setTeamMembers(members);
    } catch (error) {
      console.error('Failed to load team members:', error);
      setTeamMembers([]);
    }
  };

  const createTeam = async () => {
    if (!newTeamName.trim()) return;
    
    try {
      await api.createTeam({ name: newTeamName });
      setNewTeamName('');
      setShowNewTeamForm(false);
      loadTeams();
    } catch (error) {
      console.error('Failed to create team:', error);
    }
  };

  const inviteTeamMember = async () => {
    if (!memberEmail.trim() || !selectedTeam) return;
    
    try {
      await api.inviteTeamMember(selectedTeam.id, memberEmail.trim(), memberRole);
      setMemberEmail('');
      setMemberRole('member');
      setShowInviteMemberForm(false);
      loadTeamMembers(selectedTeam.id);
    } catch (error) {
      console.error('Failed to invite team member:', error);
      alert('Failed to invite team member. Please check if the email exists.');
    }
  };

  const openChat = (channel: Channel) => {
    router.push(`/chat/${channel.id}`);
  };

  const openTasks = (team: Team) => {
    router.push(`/tasks/${team.id}`);
  };

  if (authLoading || loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <h1 className="text-2xl font-bold text-gray-900">CBA Lite</h1>
              <div className="flex items-center space-x-2">
                <div className={`w-2 h-2 rounded-full ${wsConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <span className="text-sm text-gray-600">
                  {wsConnected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-600">
                Welcome, {user?.first_name}!
              </span>
              <button
                onClick={() => {
                  api.logout();
                  router.push('/login');
                }}
                className="text-sm text-red-600 hover:text-red-800"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow">
              <div className="p-4 border-b">
                <div className="flex items-center justify-between">
                  <h2 className="text-lg font-semibold">Teams</h2>
                  <button
                    onClick={() => setShowNewTeamForm(!showNewTeamForm)}
                    className="text-blue-600 hover:text-blue-800 text-sm"
                  >
                    + New Team
                  </button>
                </div>
                
                {showNewTeamForm && (
                  <div className="mt-4 space-y-2">
                    <input
                      type="text"
                      placeholder="Team name"
                      value={newTeamName}
                      onChange={(e) => setNewTeamName(e.target.value)}
                      className="w-full px-3 py-2 border rounded-md text-sm"
                    />
                    <div className="flex space-x-2">
                      <button
                        onClick={createTeam}
                        className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                      >
                        Create
                      </button>
                      <button
                        onClick={() => {
                          setShowNewTeamForm(false);
                          setNewTeamName('');
                        }}
                        className="px-3 py-1 bg-gray-300 text-gray-700 rounded text-sm hover:bg-gray-400"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                )}
              </div>
              
              <div className="p-4">
                {teams && teams.length > 0 ? teams.map((team) => (
                  <div
                    key={team.id}
                    onClick={() => setSelectedTeam(team)}
                    className={`p-3 rounded-md cursor-pointer mb-2 ${
                      selectedTeam?.id === team.id
                        ? 'bg-blue-100 text-blue-800'
                        : 'hover:bg-gray-100'
                    }`}
                  >
                    <div className="font-medium">{team.name}</div>
                    <div className="text-sm text-gray-600 truncate">
                      {team.description || 'No description'}
                    </div>
                  </div>
                )) : (
                  <div className="text-center text-gray-500 py-8">
                    <p>No teams found.</p>
                    <p className="text-sm">Create your first team to get started!</p>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Main Content */}
          <div className="lg:col-span-3">
            {selectedTeam ? (
              <div className="bg-white rounded-lg shadow">
                <div className="p-4 border-b">
                  <h2 className="text-xl font-semibold">{selectedTeam.name}</h2>
                  <p className="text-gray-600">{selectedTeam.description}</p>
                  
                  <div className="flex space-x-4 mt-4">
                    <button
                      onClick={() => setActiveTab('chat')}
                      className={`px-4 py-2 rounded-md ${
                        activeTab === 'chat'
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                      }`}
                    >
                      Channels ({channels.length})
                    </button>
                    <button
                      onClick={() => setActiveTab('tasks')}
                      className={`px-4 py-2 rounded-md ${
                        activeTab === 'tasks'
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                      }`}
                    >
                      Tasks ({tasks.length})
                    </button>
                    <button
                      onClick={() => setActiveTab('members')}
                      className={`px-4 py-2 rounded-md ${
                        activeTab === 'members'
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                      }`}
                    >
                      Members ({teamMembers.length})
                    </button>
                  </div>
                </div>

                <div className="p-4">
                  {activeTab === 'chat' && (
                    <div className="space-y-4">
                      <div className="flex justify-between items-center">
                        <h3 className="text-lg font-medium">Channels</h3>
                        <button
                          onClick={() => router.push(`/channels/${selectedTeam.id}/new`)}
                          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                        >
                          + New Channel
                        </button>
                      </div>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {channels && channels.length > 0 ? channels.map((channel) => (
                          <div
                            key={channel.id}
                            className="p-4 border rounded-lg hover:shadow-md transition-shadow cursor-pointer"
                            onClick={() => openChat(channel)}
                          >
                            <div className="flex items-center space-x-2">
                              <span className="text-gray-500">
                                {channel.is_private ? '🔒' : '#'}
                              </span>
                              <h4 className="font-medium">{channel.name}</h4>
                            </div>
                            <p className="text-sm text-gray-600 mt-1">
                              {channel.description || 'No description'}
                            </p>
                            <div className="text-xs text-gray-500 mt-2">
                              {channel.type} channel
                            </div>
                          </div>
                        )) : (
                        <div className="text-center text-gray-500 py-12">
                          <p>No channels found.</p>
                          <p className="text-sm">Create your first channel to get started!</p>
                        </div>
                      )}
                      </div>
                    </div>
                  )}

                  {activeTab === 'tasks' && (
                    <div className="space-y-4">
                      <div className="flex justify-between items-center">
                        <h3 className="text-lg font-medium">Tasks</h3>
                        <button
                          onClick={() => openTasks(selectedTeam)}
                          className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
                        >
                          Manage Tasks
                        </button>
                      </div>
                      
                      <div className="space-y-3">
                        {tasks && tasks.length > 0 ? tasks.slice(0, 5).map((task) => (
                          <div
                            key={task.id}
                            className="p-4 border rounded-lg hover:shadow-md transition-shadow"
                          >
                            <div className="flex items-start justify-between">
                              <div className="flex-1">
                                <h4 className="font-medium">{task.title}</h4>
                                <p className="text-sm text-gray-600 mt-1">
                                  {task.description}
                                </p>
                                <div className="flex items-center space-x-4 mt-2">
                                  <span className={`px-2 py-1 rounded text-xs ${
                                    task.status === 'done' ? 'bg-green-100 text-green-800' :
                                    task.status === 'in_progress' ? 'bg-yellow-100 text-yellow-800' :
                                    'bg-gray-100 text-gray-800'
                                  }`}>
                                    {task.status.replace('_', ' ')}
                                  </span>
                                  <span className={`px-2 py-1 rounded text-xs ${
                                    task.priority === 'urgent' ? 'bg-red-100 text-red-800' :
                                    task.priority === 'high' ? 'bg-orange-100 text-orange-800' :
                                    task.priority === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                                    'bg-gray-100 text-gray-800'
                                  }`}>
                                    {task.priority}
                                  </span>
                                </div>
                              </div>
                            </div>
                          </div>
                        )) : (
                        <div className="text-center text-gray-500 py-12">
                          <p>No tasks found.</p>
                          <p className="text-sm">Create your first task to get organized!</p>
                        </div>
                      )}
                      </div>
                    </div>
                  )}

                  {activeTab === 'members' && (
                    <div className="space-y-4">
                      <div className="flex justify-between items-center">
                        <h3 className="text-lg font-medium">Team Members</h3>
                        <button
                          onClick={() => setShowInviteMemberForm(!showInviteMemberForm)}
                          className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700"
                        >
                          + Invite Member
                        </button>
                      </div>

                      {showInviteMemberForm && (
                        <div className="bg-gray-50 p-4 rounded-lg border">
                          <h4 className="text-md font-medium mb-3">Invite New Member</h4>
                          <div className="space-y-3">
                            <div>
                              <label className="block text-sm font-medium text-gray-700 mb-1">
                                Email Address
                              </label>
                              <input
                                type="email"
                                placeholder="Enter email address"
                                value={memberEmail}
                                onChange={(e) => setMemberEmail(e.target.value)}
                                className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                              />
                            </div>
                            <div>
                              <label className="block text-sm font-medium text-gray-700 mb-1">
                                Role
                              </label>
                              <select
                                value={memberRole}
                                onChange={(e) => setMemberRole(e.target.value)}
                                className="w-full px-3 py-2 border rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                              >
                                <option value="member">Member</option>
                                <option value="admin">Admin</option>
                              </select>
                            </div>
                            <div className="flex space-x-2">
                              <button
                                onClick={inviteTeamMember}
                                disabled={!memberEmail.trim()}
                                className="px-4 py-2 bg-purple-600 text-white rounded text-sm hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
                              >
                                Send Invite
                              </button>
                              <button
                                onClick={() => {
                                  setShowInviteMemberForm(false);
                                  setMemberEmail('');
                                  setMemberRole('member');
                                }}
                                className="px-4 py-2 bg-gray-300 text-gray-700 rounded text-sm hover:bg-gray-400"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        </div>
                      )}
                      
                      <div className="space-y-3">
                        {teamMembers && teamMembers.length > 0 ? teamMembers.map((member) => (
                          <div
                            key={member.user_id}
                            className="p-4 border rounded-lg hover:shadow-md transition-shadow"
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-3">
                                <div className="w-10 h-10 bg-purple-500 rounded-full flex items-center justify-center text-white font-medium">
                                  {member.user.first_name?.[0]?.toUpperCase() || 'U'}
                                </div>
                                <div>
                                  <h4 className="font-medium">
                                    {member.user.first_name} {member.user.last_name}
                                  </h4>
                                  <p className="text-sm text-gray-600">@{member.user.username}</p>
                                  <p className="text-xs text-gray-500">{member.user.email}</p>
                                </div>
                              </div>
                              <div className="text-right">
                                <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${
                                  member.role === 'owner' ? 'bg-yellow-100 text-yellow-800' :
                                  member.role === 'admin' ? 'bg-blue-100 text-blue-800' :
                                  'bg-gray-100 text-gray-800'
                                }`}>
                                  {member.role}
                                </span>
                                <p className="text-xs text-gray-500 mt-1">
                                  Joined {new Date(member.joined_at).toLocaleDateString()}
                                </p>
                              </div>
                            </div>
                          </div>
                        )) : (
                          <div className="text-center text-gray-500 py-12">
                            <p>No team members found.</p>
                            <p className="text-sm">Invite members to start collaborating!</p>
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ) : (
              <div className="bg-white rounded-lg shadow p-12 text-center">
                <h2 className="text-xl font-semibold text-gray-600 mb-2">
                  Select a team to get started
                </h2>
                <p className="text-gray-500">
                  Choose a team from the sidebar to view channels and tasks
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}