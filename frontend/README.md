# CBA Lite Frontend

A modern, responsive frontend for the CBA Lite real-time team collaboration platform built with Next.js, React, TypeScript, and Tailwind CSS.

## Features

- **Authentication**: Login/register with JWT token management
- **Real-time Chat**: WebSocket-powered messaging with typing indicators  
- **Team Management**: Create and manage teams with channels
- **Task Management**: Kanban-style task board with drag-and-drop functionality
- **Dashboard**: Overview of teams, channels, and recent tasks
- **Responsive Design**: Works seamlessly on desktop and mobile devices

## Technology Stack

- **Next.js 15** - React framework with App Router
- **React 19** - UI library with latest features
- **TypeScript** - Type-safe development
- **Tailwind CSS 4** - Modern utility-first CSS framework
- **WebSocket** - Real-time communication
- **Context API** - State management for authentication

## Getting Started

### Prerequisites

- Node.js 18+ 
- Backend API running on `http://localhost:8080`

### Installation

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables:
```bash
# .env.local is already configured with:
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1/ws
```

3. Run the development server:
```bash
npm run dev
```

4. Open [http://localhost:3000](http://localhost:3000) in your browser

## Project Structure

```
frontend/
├── src/
│   ├── app/              # Next.js App Router pages
│   │   ├── login/        # Authentication pages  
│   │   ├── register/
│   │   ├── dashboard/    # Main dashboard
│   │   ├── chat/         # Real-time chat interface
│   │   └── tasks/        # Task management
│   ├── contexts/         # React Context providers
│   │   └── AuthContext.tsx
│   ├── lib/              # Utilities and API clients
│   │   ├── api.ts        # Backend API client
│   │   └── websocket.ts  # WebSocket client
│   └── components/       # Reusable UI components
├── public/               # Static assets
└── package.json
```

## API Integration

The frontend integrates with the backend API using:

### HTTP Endpoints
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/teams` - Get user teams
- `GET /api/v1/teams/{id}/channels` - Get team channels
- `POST /api/v1/channels/{id}/messages` - Send messages
- `GET /api/v1/teams/{id}/tasks` - Get team tasks
- `POST /api/v1/teams/{id}/tasks` - Create new tasks

### WebSocket Connection
- Real-time messaging
- Typing indicators  
- Task updates
- User presence

## Key Features

### Authentication
- JWT token-based authentication
- Automatic token refresh
- Protected routes with context
- Login/logout functionality

### Dashboard
- Team overview with channels and tasks
- WebSocket connection status indicator
- Quick navigation to chat and task management
- Team creation and management

### Real-time Chat  
- Channel-based messaging
- Real-time message delivery via WebSocket
- Typing indicators
- Message history
- Responsive chat interface

### Task Management
- Kanban-style board (Todo, In Progress, Review, Done)
- Task creation with priority and due dates
- Status updates with visual feedback  
- Priority and status color coding
- Task filtering and organization

## Usage

1. **Start the backend server** (see backend README)
2. **Register a new account** or login with existing credentials
3. **Create a team** from the dashboard
4. **Create channels** within your team for different topics
5. **Start chatting** in real-time with team members
6. **Create and manage tasks** using the Kanban board
7. **Invite team members** via email to collaborate

## Scripts

```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server  
npm run lint         # Run ESLint
```

## Environment Variables

The frontend is pre-configured to connect to:
- **API**: `http://localhost:8080/api/v1`
- **WebSocket**: `ws://localhost:8080/api/v1/ws`

Update `.env.local` if your backend runs on different ports.

## License

Private and Confidential