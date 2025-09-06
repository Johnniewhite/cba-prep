'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import api from '@/lib/api';

interface Task {
  id: string;
  team_id: string;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'review' | 'done' | 'cancelled';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  assignee_id?: string;
  created_by: string;
  due_date?: string;
  created_at: string;
  tags?: string[];
}

export default function TasksPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const teamId = params.teamId as string;
  
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [showNewTaskForm, setShowNewTaskForm] = useState(false);
  const [newTask, setNewTask] = useState({
    title: '',
    description: '',
    priority: 'medium' as const,
    due_date: '',
  });

  useEffect(() => {
    if (!user) {
      router.push('/login');
      return;
    }

    if (teamId) {
      loadTasks();
    }
  }, [teamId, user]);

  const loadTasks = async () => {
    try {
      const tasksData = await api.getTasks(teamId);
      setTasks(tasksData);
    } catch (error) {
      console.error('Failed to load tasks:', error);
    } finally {
      setLoading(false);
    }
  };

  const createTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTask.title.trim()) return;

    try {
      await api.createTask(teamId, {
        ...newTask,
        due_date: newTask.due_date || undefined,
      });
      
      setNewTask({
        title: '',
        description: '',
        priority: 'medium',
        due_date: '',
      });
      setShowNewTaskForm(false);
      loadTasks();
    } catch (error) {
      console.error('Failed to create task:', error);
    }
  };

  const updateTaskStatus = async (taskId: string, status: Task['status']) => {
    try {
      await api.updateTask(taskId, { status });
      loadTasks();
    } catch (error) {
      console.error('Failed to update task status:', error);
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'urgent':
        return 'bg-red-100 text-red-800';
      case 'high':
        return 'bg-orange-100 text-orange-800';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800';
      case 'low':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'done':
        return 'bg-green-100 text-green-800';
      case 'in_progress':
        return 'bg-blue-100 text-blue-800';
      case 'review':
        return 'bg-purple-100 text-purple-800';
      case 'todo':
        return 'bg-gray-100 text-gray-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const groupTasksByStatus = (tasks: Task[]) => {
    return tasks.reduce((acc, task) => {
      if (!acc[task.status]) {
        acc[task.status] = [];
      }
      acc[task.status].push(task);
      return acc;
    }, {} as Record<string, Task[]>);
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-lg">Loading tasks...</div>
      </div>
    );
  }

  const groupedTasks = groupTasksByStatus(tasks);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b p-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <button
              onClick={() => router.push('/dashboard')}
              className="text-gray-500 hover:text-gray-700"
            >
              ← Back to Dashboard
            </button>
            <h1 className="text-2xl font-semibold">Task Management</h1>
          </div>
          <button
            onClick={() => setShowNewTaskForm(true)}
            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
          >
            + New Task
          </button>
        </div>
      </div>

      <div className="max-w-7xl mx-auto p-6">
        {/* New Task Form */}
        {showNewTaskForm && (
          <div className="mb-6 bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium mb-4">Create New Task</h3>
            <form onSubmit={createTask} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Task Title
                </label>
                <input
                  type="text"
                  value={newTask.title}
                  onChange={(e) => setNewTask({ ...newTask, title: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                  placeholder="Enter task title"
                  required
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <textarea
                  value={newTask.description}
                  onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                  rows={3}
                  placeholder="Enter task description"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Priority
                  </label>
                  <select
                    value={newTask.priority}
                    onChange={(e) => setNewTask({ ...newTask, priority: e.target.value as any })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                  >
                    <option value="low">Low</option>
                    <option value="medium">Medium</option>
                    <option value="high">High</option>
                    <option value="urgent">Urgent</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Due Date
                  </label>
                  <input
                    type="date"
                    value={newTask.due_date}
                    onChange={(e) => setNewTask({ ...newTask, due_date: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-green-500"
                  />
                </div>
              </div>

              <div className="flex space-x-4">
                <button
                  type="submit"
                  className="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
                >
                  Create Task
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowNewTaskForm(false);
                    setNewTask({
                      title: '',
                      description: '',
                      priority: 'medium',
                      due_date: '',
                    });
                  }}
                  className="px-6 py-2 bg-gray-300 text-gray-700 rounded-md hover:bg-gray-400"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Tasks Kanban Board */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {['todo', 'in_progress', 'review', 'done'].map((status) => (
            <div key={status} className="bg-white rounded-lg shadow">
              <div className="p-4 border-b">
                <h3 className="font-medium text-gray-900 capitalize">
                  {status.replace('_', ' ')} ({groupedTasks[status]?.length || 0})
                </h3>
              </div>
              
              <div className="p-4 space-y-3">
                {groupedTasks[status]?.map((task) => (
                  <div
                    key={task.id}
                    className="p-4 border rounded-lg hover:shadow-md transition-shadow cursor-pointer"
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h4 className="font-medium text-sm">{task.title}</h4>
                      <div className="flex space-x-1">
                        <span className={`px-2 py-1 rounded text-xs ${getPriorityColor(task.priority)}`}>
                          {task.priority}
                        </span>
                      </div>
                    </div>
                    
                    {task.description && (
                      <p className="text-sm text-gray-600 mb-3 line-clamp-2">
                        {task.description}
                      </p>
                    )}
                    
                    {task.due_date && (
                      <p className="text-xs text-gray-500 mb-2">
                        Due: {formatDate(task.due_date)}
                      </p>
                    )}

                    <div className="flex space-x-1">
                      {status !== 'todo' && (
                        <button
                          onClick={() => {
                            const prevStatus = status === 'in_progress' ? 'todo' : status === 'review' ? 'in_progress' : 'review';
                            updateTaskStatus(task.id, prevStatus as Task['status']);
                          }}
                          className="px-2 py-1 bg-gray-200 text-gray-700 rounded text-xs hover:bg-gray-300"
                        >
                          ←
                        </button>
                      )}
                      
                      {status !== 'done' && (
                        <button
                          onClick={() => {
                            const nextStatus = status === 'todo' ? 'in_progress' : status === 'in_progress' ? 'review' : 'done';
                            updateTaskStatus(task.id, nextStatus as Task['status']);
                          }}
                          className="px-2 py-1 bg-blue-200 text-blue-700 rounded text-xs hover:bg-blue-300"
                        >
                          →
                        </button>
                      )}
                    </div>
                  </div>
                ))}
                
                {(!groupedTasks[status] || groupedTasks[status].length === 0) && (
                  <div className="text-center text-gray-500 py-8 text-sm">
                    No tasks in {status.replace('_', ' ')}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        {tasks.length === 0 && (
          <div className="text-center text-gray-500 py-12">
            <h3 className="text-lg font-medium mb-2">No tasks found</h3>
            <p className="text-sm">Create your first task to get started!</p>
          </div>
        )}
      </div>
    </div>
  );
}