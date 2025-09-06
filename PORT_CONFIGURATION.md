# 🚀 Port Configuration Guide

> **Database Manager Application - Environment-Based Configuration**

## 📋 Overview

This application uses environment variables for flexible port configuration, making deployment and development easier across different environments.

## ⚙️ Configuration Files

### Backend Configuration
**File:** `backend/.env`
```env
# Server Configuration
PORT=8080                                    # Backend API port

# Security (IMPORTANT: Change in production!)
JWT_SECRET=your_jwt_secret_here_change_this

# Database Connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your_secure_password
```

### Frontend Configuration
**File:** `frontend/.env`
```env
# Backend Integration
VITE_BACKEND_PORT=8080                       # Must match backend PORT
VITE_API_BASE_URL=http://localhost:8080/api

# Frontend Configuration (automatically detected by Vite)
VITE_FRONTEND_PORT=3000                      # Frontend development port
```

## 🔧 Dynamic Configuration System

### Frontend API Configuration
**File:** `frontend/src/lib/config.js`

```javascript
// Environment-based configuration helper
export const config = {
    // Dynamically get backend port from environment
    getBackendPort: () => {
        return import.meta.env.VITE_BACKEND_PORT || '8080';
    },
    
    // Generate API base URL dynamically  
    getApiUrl: (endpoint = '') => {
        const port = import.meta.env.VITE_BACKEND_PORT || '8080';
        const baseUrl = import.meta.env.VITE_API_BASE_URL || `http://localhost:${port}/api`;
        return endpoint ? `${baseUrl}${endpoint}` : baseUrl;
    },
    
    // Get backend base URL
    getBackendUrl: () => {
        const port = import.meta.env.VITE_BACKEND_PORT || '8080';
        return `http://localhost:${port}`;
    }
};
```

### Usage in Frontend Components
```javascript
import { config } from '$lib/config.js';

// Dynamic API calls
const response = await fetch(config.getApiUrl('/database'));
const databases = await fetch(config.getApiUrl('/database-management/collections'));
```
## 🚀 Port Change Instructions

### Method 1: Environment Variables (Recommended)

1. **Change Backend Port:**
   ```bash
   # Edit backend/.env
   PORT=8081
   ```

2. **Update Frontend Configuration:**
   ```bash
   # Edit frontend/.env
   VITE_BACKEND_PORT=8081
   VITE_API_BASE_URL=http://localhost:8081/api
   ```

3. **Restart Application:**
   ```bash
   # Windows
   start-dev.bat
   
   # Linux/macOS
   ./start-dev.sh
   ```

### Method 2: Direct Environment Override

```bash
# Windows PowerShell
$env:PORT=8081; cd backend; go run main.go

# Linux/macOS
PORT=8081 cd backend && go run main.go
```

## 🔄 Port Conflict Resolution

### Common Port Conflicts

#### Backend Port 8080 in Use
```bash
# Check what's using port 8080
netstat -ano | findstr :8080    # Windows
lsof -i :8080                   # Linux/macOS

# Change to alternative port
# Edit backend/.env: PORT=8081
# Edit frontend/.env: VITE_BACKEND_PORT=8081
```

#### Frontend Port 3000 in Use
```bash
# Vite will automatically find next available port
# Or specify custom port:
# Edit frontend/.env: VITE_FRONTEND_PORT=3001
```

## 📊 Configuration Validation

### Health Check Endpoints
```bash
# Backend health check
curl http://localhost:8080/api/health

# Frontend accessibility
curl http://localhost:3000
```

### Environment Verification
```javascript
// In browser console (F12)
console.log('Backend Port:', import.meta.env.VITE_BACKEND_PORT);
console.log('API Base URL:', import.meta.env.VITE_API_BASE_URL);
```

## 🏭 Production Configuration

### Docker Deployment
```dockerfile
# Backend
ENV PORT=8080
ENV JWT_SECRET=production_secret_here

# Frontend Build
ENV VITE_BACKEND_PORT=8080
ENV VITE_API_BASE_URL=https://api.yourdomain.com/api
```

### Nginx Reverse Proxy
```nginx
# Proxy backend API
location /api/ {
    proxy_pass http://localhost:8080/api/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}

# Serve frontend
location / {
    root /var/www/frontend/dist;
    try_files $uri $uri/ /index.html;
}
```

## 🛠️ Development vs Production

### Development Mode
- Backend: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- API Base: `http://localhost:8080/api`

### Production Mode
- Backend: `https://yourdomain.com:8080` or behind reverse proxy
- Frontend: `https://yourdomain.com`
- API Base: `https://yourdomain.com/api`

## 🔍 Troubleshooting

### Issue: "Failed to fetch" errors
**Solution:**
1. Verify backend is running: `curl http://localhost:8080/api/health`
2. Check frontend .env configuration
3. Ensure ports match between backend and frontend config

### Issue: CORS errors
**Solution:**
1. Backend automatically configures CORS for frontend port
2. If using custom frontend port, update backend CORS settings
3. In production, configure proper CORS origins

### Issue: Environment variables not loading
**Solution:**
1. Restart development server after .env changes
2. Vite requires `VITE_` prefix for frontend variables
3. Backend reads .env automatically on startup

## 📝 Configuration Best Practices

1. **🔒 Security:**
   - Never commit .env files with sensitive data
   - Use strong JWT_SECRET in production
   - Change default setup passwords

2. **🚀 Performance:**
   - Use consistent port numbers across environments
   - Configure proper connection pooling for databases
   - Enable gzip compression in production

3. **🔧 Maintainability:**
   - Document port changes in team communications
   - Use descriptive environment variable names
   - Keep .env.example files updated

---

**💡 Pro Tip:** The application automatically detects environment changes. Simply update .env files and restart for instant configuration updates!

## 🆘 Quick Reference

| Service | Default Port | Config File | Environment Variable |
|---------|-------------|-------------|---------------------|
| Backend API | 8080 | `backend/.env` | `PORT` |
| Frontend Dev | 3000 | `frontend/.env` | `VITE_FRONTEND_PORT` |
| Database | 5432 | `backend/.env` | `DB_PORT` |

---

*For more detailed configuration, see the main README.md file.*

### Development
```bash
# Backend
PORT=8080

# Frontend  
VITE_API_BASE_URL=http://localhost:8080/api
```

### Production
```bash
# Backend
PORT=80

# Frontend
VITE_API_BASE_URL=https://your-domain.com/api
```

## Benefits
1. **No Hardcoded Ports**: Semua port dikonfigurasi via environment
2. **Easy Deployment**: Ganti environment variables tanpa rebuild
3. **Development Flexibility**: Developer bisa menggunakan port yang berbeda
4. **Production Ready**: Environment-based configuration untuk deployment
