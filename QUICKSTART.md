# CBA Lite - Quick Start Guide

## Prerequisites

- Docker and Docker Compose
- Node.js 18+ (for frontend)
- Go 1.21+ (for backend development)

## 🚀 Start the Full Application

### Option 1: Using Local Database (Recommended for Development)

1. **Start the Backend with Local Database:**
   ```bash
   cd backend
   make dev
   ```
   This will:
   - Start PostgreSQL on port 5433
   - Start Redis on port 6380
   - Run database migrations
   - Start the Go backend on port 8080

2. **Start the Frontend:**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   Frontend will be available at: http://localhost:3000

### Option 2: Using Docker Compose (Full Stack)

```bash
cd backend
make docker-up
```
This starts everything in containers.

### Option 3: Using Remote Database

If you want to use your remote database at `168.231.113.231:5433`:

```bash
cd backend
make dev-remote
```

## 📋 What You Can Do

1. **Register/Login** at http://localhost:3000
2. **Create a Team** from the dashboard
3. **Create Channels** within your team
4. **Chat in Real-time** with WebSocket messaging
5. **Manage Tasks** with the Kanban board
6. **Invite Team Members** via email

## 🔧 Services & Ports

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **PostgreSQL**: localhost:5433
- **Redis**: localhost:6380
- **WebSocket**: ws://localhost:8080/api/v1/ws

## 🐛 Troubleshooting

### Database Connection Issues

If you get `connection refused` errors:

1. **Check if PostgreSQL is running:**
   ```bash
   docker ps | grep postgres
   ```

2. **Test database connection:**
   ```bash
   PGPASSWORD=cbaPREP2025 psql -h localhost -p 5433 -U cba -d cbalite
   ```

3. **Restart services:**
   ```bash
   cd backend
   make stop
   make dev
   ```

### Backend Compilation Errors

If you get Go compilation errors, make sure dependencies are installed:

```bash
cd backend
make deps
```

### Frontend Issues

If frontend doesn't connect to backend:

1. Check that backend is running on port 8080
2. Verify `.env.local` has correct API URL:
   ```
   NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
   ```

## 🔄 Quick Commands

```bash
# Backend
cd backend
make dev          # Start with local database
make dev-remote   # Start with remote database  
make stop         # Stop all services
make docker-up    # Start everything with Docker

# Frontend
cd frontend
npm run dev       # Start development server
npm run build     # Build for production
```

## 📊 Health Check

Visit these URLs to verify everything is working:

- Backend Health: http://localhost:8080/api/v1/health
- Frontend: http://localhost:3000
- PostgreSQL: `PGPASSWORD=cbaPREP2025 psql -h localhost -p 5433 -U cba -d cbalite -c "SELECT version();"`



**Application:**
- Register a new account at http://localhost:3000/register
- No default users are created

