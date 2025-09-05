# Database Manager Application

Database Manager adalah aplikasi web full-stack yang memungkinkan pengguna untuk mengelola multiple database connections dan membuat REST API endpoints secara otomatis.

## Fitur

1. **User Authentication** - Login/Register dengan email dan password
2. **Multi-Database Support** - MySQL, PostgreSQL, MongoDB
3. **Database Connection Management** - Connect ke database lokal dan public
4. **Database Explorer** - Lihat collections/tables dari database yang terkoneksi
5. **Auto REST API Generation** - Generate API endpoints untuk collections/tables
6. **API Key Management** - Create dan manage API keys untuk akses endpoint
7. **Endpoint Management** - Enable/disable API endpoints
8. **API Logging** - Log semua aktivitas API requests
9. **Interactive Dashboard** - Dashboard dengan informasi database dan statistics
10. **Collection/Table Editor** - Edit struktur database (coming soon)

## Tech Stack

### Backend
- **Go Fiber** - Web framework
- **GORM** - ORM untuk database operations
- **JWT** - Authentication
- **SQLite** - Database untuk menyimpan metadata aplikasi
- **MongoDB Driver** - Koneksi ke MongoDB
- **MySQL Driver** - Koneksi ke MySQL
- **PostgreSQL Driver** - Koneksi ke PostgreSQL

### Frontend
- **Svelte/SvelteKit** - Frontend framework
- **Axios** - HTTP client
- **Chart.js** - Data visualization
- **CSS** - Styling (no additional CSS framework)

## Prerequisites

- Go 1.21+
- Node.js 18+
- npm atau yarn

## Installation

1. **Clone repository**
   ```bash
   git clone <repository-url>
   cd db-manager-app
   ```

2. **Install dependencies**
   ```bash
   npm run install:all
   ```

3. **Setup environment variables**
   
   Copy `.env.example` to `.env` di folder backend:
   ```bash
   cd backend
   cp .env.example .env
   ```
   
   Edit `.env` file:
   ```env
   PORT=8080
   JWT_SECRET=your-very-secure-jwt-secret-key
   DATABASE_URL=sqlite:./app.db
   ```

## Development

**Run both backend and frontend in development mode:**
```bash
npm run dev
```

**Run individually:**
```bash
# Backend only (port 8080)
npm run dev:backend

# Frontend only (port 3000)
npm run dev:frontend
```

## Usage

1. **Register/Login**
   - Buka http://localhost:3000
   - Daftar akun baru atau login dengan akun existing

2. **Add Database Connection**
   - Pergi ke halaman "Connections"
   - Klik "Add Connection"
   - Isi form dengan detail database
   - Test connection terlebih dahulu
   - Save connection

3. **Manage Database**
   - Klik "Manage" pada connection yang sudah dibuat
   - Lihat list collections/tables
   - Generate API endpoints untuk collection tertentu

4. **API Management**
   - Pergi ke halaman "API Management"
   - Create API keys untuk akses endpoints
   - Enable/disable endpoints
   - Monitor API logs

5. **Using Generated APIs**
   
   **Example API calls:**
   ```bash
   # GET all documents from a collection
   curl -H "X-API-Key: your-api-key" \
        http://localhost:8080/api/collection-name
   
   # POST new document
   curl -X POST \
        -H "X-API-Key: your-api-key" \
        -H "Content-Type: application/json" \
        -d '{"field": "value"}' \
        http://localhost:8080/api/collection-name
   
   # GET document by ID
   curl -H "X-API-Key: your-api-key" \
        http://localhost:8080/api/collection-name/document-id
   
   # PUT update document
   curl -X PUT \
        -H "X-API-Key: your-api-key" \
        -H "Content-Type: application/json" \
        -d '{"field": "updated-value"}' \
        http://localhost:8080/api/collection-name/document-id
   
   # DELETE document
   curl -X DELETE \
        -H "X-API-Key: your-api-key" \
        http://localhost:8080/api/collection-name/document-id
   ```

## Project Structure

```
db-manager-app/
├── backend/
│   ├── config/           # Database config
│   ├── handlers/         # HTTP handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── utils/           # Utilities (auth, etc)
│   ├── main.go          # Main application
│   ├── go.mod           # Go dependencies
│   └── .env             # Environment variables
├── frontend/
│   ├── src/
│   │   ├── lib/         # Shared utilities
│   │   ├── routes/      # SvelteKit routes
│   │   └── app.html     # HTML template
│   ├── package.json     # Frontend dependencies
│   ├── svelte.config.js # Svelte configuration
│   └── vite.config.js   # Vite configuration
├── package.json         # Root package.json
└── README.md
```

## API Documentation

### Authentication Endpoints

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `GET /api/auth/profile` - Get user profile (protected)

### Database Endpoints

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

## Database Support

### MySQL
- Default port: 3306
- Connection string: `username:password@tcp(host:port)/database`

### PostgreSQL
- Default port: 5432
- Connection string: `host=host user=username password=password dbname=database port=port sslmode=disable`

### MongoDB
- Default port: 27017
- Connection string: `mongodb://username:password@host:port/database`

## Security

- JWT tokens untuk user authentication
- API keys untuk endpoint access
- Password hashing dengan bcrypt
- SQL injection protection dengan GORM
- CORS configuration

## Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## License

This project is licensed under the MIT License.

## Support

Untuk pertanyaan atau dukungan, silakan buat issue di repository GitHub.
