# Database Manager Application 🚀

> **🔒 Secure Project** - Password protection enabled for authorized access only

Database Manager adalah aplikasi web full-stack yang memungkinkan pengguna untuk mengelola multiple database connections dan membuat REST API endpoints secara otomatis dengan performa tinggi dan keamanan terjamin.

## 🌟 Fitur

1. **🔐 User Authentication** - Login/Register dengan email dan password
2. **🗄️ Multi-Database Support** - MySQL, PostgreSQL, MongoDB
3. **🔗 Database Connection Management** - Connect ke database lokal dan public
4. **📊 Database Explorer** - Lihat collections/tables dari database yang terkoneksi
5. **⚡ Auto REST API Generation** - Generate API endpoints untuk collections/tables
6. **🔑 API Key Management** - Create dan manage API keys untuk akses endpoint
7. **🎛️ Endpoint Management** - Enable/disable API endpoints
8. **📝 API Logging** - Log semua aktivitas API requests
9. **📈 Interactive Dashboard** - Dashboard dengan informasi database dan statistics
10. **✏️ Collection/Table Editor** - Edit struktur database (coming soon)

## 🛠️ Tech Stack

### Backend
- **Go Fiber** - High-performance web framework
- **GORM** - ORM dengan memory optimization
- **JWT** - Secure authentication
- **SQLite** - Metadata storage
- **MongoDB Driver** - MongoDB connections
- **MySQL Driver** - MySQL connections  
- **PostgreSQL Driver** - PostgreSQL connections

### Frontend
- **Svelte/SvelteKit** - Reactive frontend framework
- **Vite** - Fast build tool dengan environment management
- **Environment-based Config** - Dynamic API configuration
- **Optimized Fetching** - Memory-efficient API calls

## 📋 Prerequisites

- **Go 1.21+** - Backend runtime
- **Node.js 18+** - Frontend development
- **PostgreSQL** - Database server (recommended)

## 🚀 Quick Start (Secure Setup)

### 📥 Installation

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd db-manager-app
   ```

2. **🔐 Secure Setup (Password Protected)**
   
   **Windows:**
   ```bash
   setup.bat
   ```
   
   **Linux/macOS:**
   ```bash
   chmod +x setup.sh
   ./setup.sh
   ```
   
   > **📝 Note:** You will be prompted for setup password. Contact administrator for access.

### ⚙️ Manual Configuration

If you have access authorization, configure these files:

**Backend Environment (backend/.env):**
```env
PORT=8080
JWT_SECRET=your_jwt_secret_here_change_this
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your_db_password
```

**Frontend Environment (frontend/.env):**
```env
# Backend Configuration
VITE_BACKEND_PORT=8080
VITE_API_BASE_URL=http://localhost:8080/api

# Frontend Configuration (automatically detected by Vite)
VITE_FRONTEND_PORT=3000
```

## 🏃‍♂️ Running the Application

### 🔥 Development Mode (Both Services)
```bash
# Windows
start-dev.bat

# Linux/macOS  
./start-dev.sh
```

### 🎯 Individual Services
```bash
# Backend only (port 8080)
cd backend && go run main.go

# Frontend only (port 3000)  
cd frontend && npm run dev
```

## 📖 Usage Guide

1. **🔐 Authentication**
   - Open http://localhost:3000
   - Register new account or login with existing credentials

2. **🗄️ Database Management**
   - Navigate to "Connections" page
   - Add new database connection
   - Test connection before saving
   - Manage multiple database connections

3. **⚡ API Generation**
   - Click "Manage" on created connection
   - View collections/tables list
   - Generate API endpoints for specific collections

4. **🔑 API Management**
   - Go to "API Management" page
   - Create API keys for endpoint access
   - Enable/disable endpoints
   - Monitor API logs and usage

## 🔒 Security Features

- **🔐 Password Protection** - Setup script requires authorization password
- **🎫 JWT Authentication** - Secure user session management
- **🔑 API Key System** - Controlled access to generated endpoints
- **🛡️ Password Hashing** - bcrypt encryption for user passwords
- **🚫 SQL Injection Protection** - GORM prepared statements
- **🌐 CORS Configuration** - Cross-origin request security
- **⚡ Memory Optimization** - Pointer-based operations for performance

## 🔧 Environment Configuration

The application uses environment variables for easy deployment:

### Backend Configuration (backend/.env)
```env
# Server Configuration
PORT=8080                                    # API server port

# Security
JWT_SECRET=your_jwt_secret_here_change_this  # Change this in production!

# Database Connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=your_secure_password
```

### Frontend Configuration (frontend/.env)
```env
# Backend Integration
VITE_BACKEND_PORT=8080
VITE_API_BASE_URL=http://localhost:8080/api

# Frontend Settings
VITE_FRONTEND_PORT=3000
```

## 🆘 Troubleshooting

### Common Setup Issues

#### ❌ "Failed to fetch" errors
```bash
# Check if backend is running on correct port
curl http://localhost:8080/api/health

# Verify frontend .env configuration
cat frontend/.env
```

#### ❌ Password protection failing
```bash
# Setup password: admin123secure
# If forgot, check setup.sh or setup.bat files
```

#### ❌ Go build failures
```bash
# Clean module cache
go clean -modcache

# Reinitialize modules
cd backend
rm go.mod go.sum
go mod init db-manager-backend
go mod tidy
```

#### ❌ Node.js dependency issues
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Port Conflicts

If default ports are in use:

1. **Change Backend Port:**
   ```bash
   # Edit backend/.env
   PORT=8081
   ```

2. **Update Frontend Config:**
   ```bash
   # Edit frontend/.env
   VITE_BACKEND_PORT=8081
   VITE_API_BASE_URL=http://localhost:8081/api
   ```

3. **Restart Services:**
   ```bash
   start-dev.bat  # Windows
   ./start-dev.sh # Linux/macOS
   ```

### Performance Optimization

- **Memory Usage:** Application uses pointer-based operations for minimal memory footprint
- **API Responses:** Optimized JSON marshaling with object pooling
- **Database Connections:** Connection pooling enabled for better performance
- **Frontend Caching:** Vite build optimization for production deployment

## 📊 Monitoring & Logs

### Application Logs
```bash
# Backend logs (stdout)
cd backend && ./db-manager-backend.exe

# Frontend development logs
cd frontend && npm run dev
```

### API Usage Monitoring
- Access "API Management" page in application
- View real-time API call logs
- Monitor endpoint usage statistics
- Track authentication attempts

## 🚀 Production Deployment

### Environment Preparation
1. **Change default passwords and secrets**
2. **Configure production database**
3. **Set secure JWT_SECRET**
4. **Enable HTTPS (recommended)**

### Build Commands
```bash
# Build frontend for production
cd frontend && npm run build

# Build backend binary
cd backend && go build -o db-manager-backend main.go
```

## 📋 Project Structure

```
db-manager-app/
├── 🔧 Configuration Files
│   ├── setup.bat                    # Windows setup (password protected)
│   ├── setup.sh                     # Linux/macOS setup (password protected)
│   ├── start-dev.bat                # Development startup (Windows)
│   ├── package.json                 # Root dependencies
│   └── README.md                    # This documentation
├── 🖥️ Backend (Go/Fiber)
│   ├── backend/.env                 # Backend environment configuration
│   ├── config/database.go           # Database configuration
│   ├── handlers/                    # HTTP request handlers
│   │   ├── api.go                   # Generated API endpoints
│   │   ├── auth.go                  # Authentication handlers
│   │   ├── database.go              # Database management
│   │   └── dynamic_api_optimized.go # Memory-optimized API generation
│   ├── models/models.go             # Data structures
│   ├── services/database.go         # Business logic
│   ├── utils/auth.go                # Authentication utilities
│   └── main.go                      # Application entry point
└── 🎨 Frontend (Svelte/SvelteKit)
    ├── frontend/.env                # Frontend environment configuration
    ├── src/lib/
    │   ├── config.js                # Environment-based API configuration
    │   ├── api.js                   # API client functions
    │   └── components/              # Reusable UI components
    ├── src/routes/                  # Application pages
    │   ├── dashboard/               # Main dashboard
    │   ├── connections/             # Database connections
    │   ├── database-management/     # Database explorer
    │   └── api-management/          # API endpoint management
    └── package.json                 # Frontend dependencies
```

## 📚 API Documentation

### Authentication Endpoints

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `GET /api/auth/profile` - Get user profile (protected)

### Database Management Endpoints

- `POST /api/database/test` - Test database connection
- `POST /api/database` - Create new connection (protected)
- `GET /api/database` - List user connections (protected)
- `GET /api/database/:id/info` - Get database info (protected)
- `DELETE /api/database/:id` - Delete connection (protected)

### API Management Endpoints

- `POST /api/api-management/keys` - Create API key (protected)
- `GET /api/api-management/keys` - List API keys (protected)
- `PUT /api/api-management/keys/:id/toggle` - Toggle API key (protected)
- `POST /api/api-management/endpoints` - Create endpoint (protected)
- `GET /api/api-management/endpoints` - List endpoints (protected)
- `PUT /api/api-management/endpoints/:id/toggle` - Toggle endpoint (protected)
- `GET /api/api-management/logs` - Get API logs (protected)

### Generated API Endpoints

All generated endpoints require `X-API-Key` header:

- `GET /api/:collection` - List all documents
- `POST /api/:collection` - Create new document
- `GET /api/:collection/:id` - Get document by ID
- `PUT /api/:collection/:id` - Update document
- `DELETE /api/:collection/:id` - Delete document

## 💡 Usage Examples

### API Examples
```bash
# GET all documents from a collection
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/users

# POST new document
curl -X POST \
     -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     -d '{"name": "John Doe", "email": "john@example.com"}' \
     http://localhost:8080/api/users

# GET document by ID
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/users/12345

# PUT update document
curl -X PUT \
     -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     -d '{"name": "Jane Doe", "email": "jane@example.com"}' \
     http://localhost:8080/api/users/12345

# DELETE document
curl -X DELETE \
     -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/users/12345
```

## 🗄️ Database Support

### MySQL
- **Default port:** 3306
- **Connection string:** `username:password@tcp(host:port)/database`
- **Features:** Full CRUD operations, table management

### PostgreSQL
- **Default port:** 5432
- **Connection string:** `host=host user=username password=password dbname=database port=port sslmode=disable`
- **Features:** Advanced queries, JSON support

### MongoDB
- **Default port:** 27017
- **Connection string:** `mongodb://username:password@host:port/database`
- **Features:** Document operations, aggregation pipelines

## 🔒 Security Considerations

### Production Security Checklist
- [ ] Change default setup password (`admin123secure`)
- [ ] Update JWT_SECRET with strong random value
- [ ] Use HTTPS in production
- [ ] Configure firewall rules
- [ ] Enable database authentication
- [ ] Regularly rotate API keys
- [ ] Monitor API usage logs
- [ ] Set up rate limiting

## 🤝 Contributing

1. **Fork repository**
2. **Create feature branch** (`git checkout -b feature/amazing-feature`)
3. **Commit changes** (`git commit -m 'Add amazing feature'`)
4. **Push to branch** (`git push origin feature/amazing-feature`)
5. **Open Pull Request**

### Development Guidelines
- Follow Go best practices for backend
- Use Svelte conventions for frontend
- Add tests for new features
- Update documentation
- Ensure security considerations

## 📞 Support & Contact

### Getting Help
- **📖 Documentation:** This README covers most use cases
- **🐛 Bug Reports:** Create GitHub issue with detailed description
- **💡 Feature Requests:** Open GitHub issue with enhancement label
- **🔒 Security Issues:** Contact maintainers privately

### Setup Password Access
> **🔑 For authorized users only:** Setup password is `admin123secure`

## 📄 License

This project is licensed under the **MIT License** - see the LICENSE file for details.

---

**🚀 Database Manager Application** - Powerful, secure, and optimized database management with auto-generated REST APIs.

*Built with ❤️ using Go, Fiber, Svelte, and modern web technologies.*
