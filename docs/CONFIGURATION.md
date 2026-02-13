# Configuration Reference
**Team Project**

Complete reference for all configuration options, environment variables, and settings.

## Table of Contents
1. [Environment Variables](#environment-variables)
2. [Backend Configuration](#backend-configuration)
3. [Frontend Configuration](#frontend-configuration)
4. [Database Configuration](#database-configuration)
5. [Feature Flags](#feature-flags)
6. [Environment-Specific Settings](#environment-specific-settings)
7. [Examples](#examples)

---

## Environment Variables

### Required Variables

These must be set for the application to run.

#### Backend (.env file)

| Variable | Example | Description | Required |
|----------|---------|-------------|----------|
| `FLASK_ENV` | `development` | Flask environment (development, staging, production) | Yes |
| `FLASK_DEBUG` | `1` | Enable Flask debug mode (0 or 1) | No (default: 0) |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string | Yes |
| `MONGO_DB` | `cloud_native` | Database name | Yes |
| `SECRET_KEY` | `your-secret-key-here` | Flask secret key for sessions | Yes |
| `LOG_LEVEL` | `INFO` | Logging level (DEBUG, INFO, WARNING, ERROR) | No (default: INFO) |

#### Frontend (.env file / .env.local)

| Variable | Example | Description | Required |
|----------|---------|-------------|----------|
| `VITE_API_BASE_URL` | `http://localhost:5001` | Backend API base URL | Yes |
| `VITE_APP_NAME` | `Powder Wireless Hardware Checkout` | Application name | No |
| `VITE_DEBUG` | `true` | Enable debug logging | No (default: false) |

---

## Backend Configuration

### Flask Configuration (app/config.py)

```python
class Config:
    """Base configuration."""
    
    # Flask
    FLASK_ENV = os.getenv('FLASK_ENV', 'development')
    DEBUG = os.getenv('FLASK_DEBUG', '0') == '1'
    TESTING = False
    
    # Security
    SECRET_KEY = os.getenv('SECRET_KEY', 'dev-key-change-in-production')
    
    # Database
    MONGO_URI = os.getenv('MONGO_URI', 'mongodb://localhost:27017')
    MONGO_DB = os.getenv('MONGO_DB', 'cloud_native')
    
    # Logging
    LOG_LEVEL = os.getenv('LOG_LEVEL', 'INFO')
    LOG_FORMAT = '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    
    # Pagination
    ITEMS_PER_PAGE = 20
    MAX_ITEMS_PER_PAGE = 100
    
    # Timeouts
    DB_TIMEOUT = 5  # seconds
    REQUEST_TIMEOUT = 30  # seconds
    
    # Security
    PASSWORD_MIN_LENGTH = 8
    PASSWORD_REQUIRE_UPPERCASE = True
    PASSWORD_REQUIRE_NUMBERS = True
    PASSWORD_REQUIRE_SPECIAL = False

class DevelopmentConfig(Config):
    """Development configuration."""
    DEBUG = True
    TESTING = False
    MONGO_URI = 'mongodb://localhost:27017'
    
class StagingConfig(Config):
    """Staging configuration."""
    DEBUG = False
    TESTING = False
    
class ProductionConfig(Config):
    """Production configuration."""
    DEBUG = False
    TESTING = False
    # Production values from environment variables
    
class TestingConfig(Config):
    """Testing configuration."""
    TESTING = True
    MONGO_DB = 'test_cloud_native'
```

### Server Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `HOST` | `0.0.0.0` | Server bind address |
| `PORT` | `5001` | Server port |
| `WORKERS` | `4` | Number of worker processes (production) |
| `TIMEOUT` | `60` | Worker timeout in seconds |

---

## Frontend Configuration

### Vite Configuration (vite.config.ts)

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  
  server: {
    port: 5173,
    strictPort: false,
    proxy: {
      '/api': {
        target: process.env.VITE_API_BASE_URL || 'http://localhost:5001',
        changeOrigin: true,
      },
    },
  },
  
  build: {
    outDir: 'dist',
    sourcemap: false,  // Set to true for debugging
    minify: 'terser',
  },
  
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
})
```

### TypeScript Configuration (tsconfig.json)

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "strict": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

---

## Database Configuration

### MongoDB Settings

#### Connection String Format

```
mongodb://[username:password@]host[:port]/[database][?options]
```

#### Examples

**Local Development:**
```
mongodb://localhost:27017/hardware_checkout
```

**Remote (Atlas):**
```
mongodb+srv://user:password@cluster.mongodb.net/hardware_checkout?retryWrites=true&w=majority
```

**Replica Set:**
```
mongodb://host1:27017,host2:27017,host3:27017/hardware_checkout?replicaSet=rs0
```

### Collection Configuration

#### Users Collection
```javascript
db.createCollection('users', {
  schema: {
    bsonType: 'object',
    required: ['userId', 'password'],
    properties: {
      _id: { bsonType: 'objectId' },
      userId: { bsonType: 'string' },  // Encrypted per SR3
      password: { bsonType: 'string' },  // Encrypted per SR3
      createdAt: { bsonType: 'date' },
      updatedAt: { bsonType: 'date' }
    }
  }
})

// Create unique index on userId (encrypted value)
db.users.createIndex({ userId: 1 }, { unique: true })
```

#### Projects Collection
```javascript
db.createCollection('projects', {
  schema: {
    bsonType: 'object',
    required: ['projectId', 'name', 'ownerUserId'],
    properties: {
      _id: { bsonType: 'objectId' },
      projectId: { bsonType: 'string' },
      name: { bsonType: 'string' },
      description: { bsonType: 'string' },
      ownerUserId: { bsonType: 'string' },
      assignedUsers: { 
        bsonType: 'array',
        items: { bsonType: 'string' }
      },
      createdAt: { bsonType: 'date' }
    }
  }
})

// TODO: Clarify owner vs assignedUsers model
db.projects.createIndex({ projectId: 1 }, { unique: true })
```

#### Hardware Sets Collection
```javascript
db.createCollection('hardware_sets', {
  schema: {
    bsonType: 'object',
    required: ['hardwareId', 'hardwareName', 'capacity'],
    properties: {
      _id: { bsonType: 'objectId' },
      hardwareId: { bsonType: 'string' },
      hardwareName: { bsonType: 'string' },
      capacity: { bsonType: 'int' },
      available: { bsonType: 'int' },
      projectAllotments: {
        bsonType: 'array',
        items: {
          bsonType: 'object',
          properties: {
            projectId: { bsonType: 'string' },
            checkedOut: { bsonType: 'int' }
          }
        }
      },
      createdAt: { bsonType: 'date' }
    }
  }
})

db.hardware_sets.createIndex({ hardwareId: 1 }, { unique: true })
```

#### Allocations Collection
```javascript
db.createCollection('allocations', {
  schema: {
    bsonType: 'object',
    required: ['projectId', 'hardwareSet', 'units', 'type'],
    properties: {
      _id: { bsonType: 'objectId' },
      projectId: { bsonType: 'string' },
      userId: { bsonType: 'string' },
      hardwareSet: { bsonType: 'string' },
      units: { bsonType: 'int' },
      type: { enum: ['checkout', 'checkin'] },
      timestamp: { bsonType: 'date' }
    }
  }
})

// Index for quick lookups by project
db.allocations.createIndex({ projectId: 1, timestamp: -1 })
```

---

## Feature Flags

**TODO:** Implement feature flag system

Proposed feature flags:

| Flag | Purpose | Default |
|------|---------|---------|
| `FEATURE_HARDWARE_CHECKOUT` | Enable hardware checkout functionality | `true` |
| `FEATURE_PROJECT_MANAGEMENT` | Enable project management | `true` |
| `FEATURE_REQUEST_APPROVAL` | Require approval for requests | `false` |
| `FEATURE_HARDWARE_ALLOCATION_LIMITS` | Enforce per-project allocation limits | `true` |
| `FEATURE_USER_PROFILES` | Enable user profile pages | `true` |
| `FEATURE_EMAIL_NOTIFICATIONS` | Send email notifications | `false` |

**Implementation Example:**
```python
# In config
FEATURES = {
    'HARDWARE_CHECKOUT': os.getenv('FEATURE_HARDWARE_CHECKOUT', 'true').lower() == 'true',
    'PROJECT_MANAGEMENT': os.getenv('FEATURE_PROJECT_MANAGEMENT', 'true').lower() == 'true',
}

# In routes
if current_app.config['FEATURES']['HARDWARE_CHECKOUT']:
    # Checkout code
    pass
else:
    return {'error': 'Feature disabled'}, 403
```

---

## Environment-Specific Settings

### Development
```bash
FLASK_ENV=development
FLASK_DEBUG=1
MONGO_URI=mongodb://localhost:27017
MONGO_DB=cloud_native_dev
SECRET_KEY=dev-secret-key
LOG_LEVEL=DEBUG
```

### Staging
```bash
FLASK_ENV=staging
FLASK_DEBUG=0
MONGO_URI=mongodb+srv://user:pass@staging-cluster.mongodb.net
MONGO_DB=cloud_native_staging
SECRET_KEY=<random-staging-key>
LOG_LEVEL=INFO
```

### Production
```bash
FLASK_ENV=production
FLASK_DEBUG=0
MONGO_URI=mongodb+srv://user:pass@prod-cluster.mongodb.net
MONGO_DB=cloud_native_prod
SECRET_KEY=<strong-random-key>
LOG_LEVEL=WARNING
```

---

## Examples

### Complete Backend .env File

```bash
# Flask Configuration
FLASK_ENV=development
FLASK_DEBUG=1

# Database
MONGO_URI=mongodb://localhost:27017
MONGO_DB=cloud_native

# Security
SECRET_KEY=change-this-in-production

# Logging
LOG_LEVEL=DEBUG

# Feature Flags
FEATURE_HARDWARE_CHECKOUT=true
FEATURE_REQUEST_APPROVAL=false
FEATURE_EMAIL_NOTIFICATIONS=false
```

### Complete Frontend .env File

```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:5001

# App Configuration
VITE_APP_NAME=Powder Wireless Hardware Checkout
VITE_DEBUG=true
```

### Docker Environment Variables

**docker-compose.yml:**
```yaml
services:
  backend:
    environment:
      - FLASK_ENV=development
      - MONGO_URI=mongodb://mongo:27017
      - MONGO_DB=cloud_native
      - SECRET_KEY=dev-key
      - LOG_LEVEL=DEBUG
      
  frontend:
    environment:
      - VITE_API_BASE_URL=http://backend:5001
      - VITE_DEBUG=true
      
  mongo:
    environment:
      - MONGO_INITDB_DATABASE=hardware_checkout
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
```

---

## Password Policy

| Setting | Requirement | Reason |
|---------|-------------|--------|
| Minimum Length | 8 characters | Balance security and usability |
| Uppercase Letters | Required | Increase entropy |
| Lowercase Letters | Implicit | Natural usage |
| Numbers | Required | Increase entropy |
| Special Characters | Optional | User preference |
| Expiration | None (for now) | Balance security and UX |

**TODO:** Determine final password policy

---

## Session & Authentication

| Setting | Value | Notes |
|---------|-------|-------|
| Session Timeout | 24 hours | TODO: Determine actual value |
| Token Type | JWT | Consider switching to session-based |
| Token Expiration | 24 hours | TODO: Implement refresh tokens |
| HTTPS Required | Development: No, Production: Yes | Enforce in production |
| CORS Allowed Origins | `http://localhost:5173` (dev) | Restrict in production |

---

## Performance Settings

| Setting | Development | Production |
|---------|-------------|-----------|
| Database Connection Pool | 10 | 50 |
| Request Timeout | 30s | 10s |
| Page Size (pagination) | 20 | 50 |
| Cache TTL | 60s | 300s |
| Compression | Off | Gzip ON |

---

## Troubleshooting Configuration

### Issue: Cannot connect to MongoDB
```bash
# Check connection string
echo $MONGO_URI

# Test connection
mongosh "$MONGO_URI"

# Verify database exists
mongosh "$MONGO_URI" --eval "db.adminCommand('ping')"
```

### Issue: CORS errors in frontend
```bash
# Check VITE_API_BASE_URL
echo $VITE_API_BASE_URL

# Verify backend is running
curl http://localhost:5001/api/health
```

### Issue: Feature flag not working
```bash
# Check feature is set
echo $FEATURE_HARDWARE_CHECKOUT

# Verify in code
# Use exact string name as in FEATURES dict
```

---

**Last Updated:** February 13, 2026  
**Version:** 1.1  
**Status:** Active
