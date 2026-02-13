# Software Design Document (SDD)
## Team Project

**Project Name:** Team Project
**Version:** 1.1.0
**Date:** February 2026
**Status:** In Development
**Course:** ECE 382V: Cloud Native App Development  

---

## Table of Contents

1. [Document Overview](#1-document-overview)
2. [Record History](#2-record-history)
3. [Stakeholders](#3-stakeholders)
4. [System Architecture](#4-system-architecture)
5. [Backend Design](#5-backend-design)
6. [Frontend Design](#6-frontend-design)
7. [Data Model Design](#7-data-model-design)
8. [UML Class Diagrams](#8-uml-class-diagrams)
9. [API Design](#9-api-design)
10. [Security Design](#10-security-design)
11. [Business Rules](#11-business-rules)
12. [Design Patterns & Workflows](#12-design-patterns--workflows)
13. [Testing Strategy](#13-testing-strategy)
14. [Implementation Status](#14-implementation-status)
15. [Technology Stack](#15-technology-stack)
16. [Future Enhancements](#16-future-enhancements)

---

## 1. Document Overview

This Software Design Document provides a detailed technical design for the Hardware-as-a-Service (HaaS) Proof of Concept application. It covers system architecture, component design, data models, and API specifications. This document references the Software Requirements Specification (SRS) for functional and non-functional requirements.

### References
- [Software Requirements Specification (SRS)](./SRS.md)
- [Requirements Traceability Matrix](./REQUIREMENTS_TRACEABILITY.md)

---

## 2. Record History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | Feb 10, 2026 | Casey Webb | Initial SDD creation with architecture, backend, frontend, and API design |
| 1.1.0 | Feb 10, 2026 | Casey Webb | Added stakeholders, record history, UML diagrams, flowcharts, security design, and testing strategy |

---

## 3. Stakeholders

### 3.1 Development Team

| Role | Name | Responsibilities |
|------|------|------------------|
| **Project Lead / Lead Developer** | Casey Webb | Overall architecture, backend development, MongoDB design, technical decisions |
| **Frontend Developer** | | 
| **Team Member** | [TBD] |  |
| **Team Member** | Anita Woodford | Documentation, 

Needs Assignments: Hardware module implementation, API endpoints, testing deployment configuration, DevOps |
### 3.2 Stakeholder Communication Matrix

| Stakeholder | Communication Frequency | Primary Topics | Responsible Party |
|------------|------------------------|-----------------|--------------------|
| Project Lead | Daily | Architecture decisions, blockers, milestones | Development Team |
| Frontend Developer | Daily | API contract, UI requirements, integration | Frontend/Backend Leads |
| Backend Developers | Daily | Database schema, API design, build tasks | Backend Lead |
| Course Instructor | Weekly | Progress, design decisions, requirements alignment | Project Lead |
| Department (ECE) | Monthly | Feasibility, resource needs, project status | Project Lead |

### 3.3 Decision Authority Matrix

| Decision Type | Authority | Approval Required? |
|---------------|-----------|-------------------|
| Architecture changes | Project Lead | Yes |
| API endpoint changes | Backend Lead + Frontend Lead | Yes |
| Database schema changes | Project Lead | Yes |
| UI/UX changes | Frontend Lead | Internal review only |
| Technology stack changes | Project Lead | Yes |
| Timeline adjustments | Project Lead + Instructor | Yes |
| Bug fixes and minor features | Development Team | No |

---

## 4. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Browser / Client                        │
│                  (Modern Browser - ES2020+)                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP/HTTPS
                            │ JSON
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Frontend Application                         │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ React 19.2.0 + TypeScript 5.9+ + Vite                    │   │
│  │ - React Router for page-based routing                    │   │
│  │ - React Context for client-side state management         │   │
│  │ - Axios for HTTP requests with interceptors              │   │
│  │ - Material-UI for component library                      │   │
│  │ - Protected Routes (ProtectedRoute.tsx)                  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                    (Port 5173)                                  │
│  - src/pages/: Auth, Home, Account, Projects, Hardware          │
│  - src/api/: http client, users, projects, hardware modules     │
│  - src/auth/: Context, Provider, hooks for authentication       │
│  - src/layouts/: AppLayout with header/footer                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │ REST API
                            │ /api/*
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Backend Application                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ Flask 3.1.2 + Python 3.12+                               │   │
│  │ - Blueprints for modular route organization              │   │
│  │ - Flask-CORS for cross-origin requests                   │   │
│  │ - PyMongo for database operations                        │   │
│  │ - No server-side sessions (stateless backend)            │   │
│  └──────────────────────────────────────────────────────────┘   │
│                    (Port 5001)                                  │
│  - app/routes/auth.py: Authentication & encryption              │
│  - app/routes/users.py: User CRUD operations                    │
│  - app/routes/projects.py: Project CRUD operations              │
│  - app/routes/hardware.py: Hardware management (to build)       │
│  - app/db.py: MongoDB connection management                     │
│  - app/mongo_utils.py: Database utilities                       │
│  - app/config.py: Configuration management                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │ MongoDB Wire Protocol
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Database Layer                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ MongoDB 7.0+                                             │   │
│  │ - users collection: Account & credentials                │   │
│  │ - projects collection: Project definitions               │   │
│  │ - hardware_sets collection: Hardware inventory           │   │
│  │ - resource_requests collection: Request tracking         │   │
│  │ - allocations collection: Checkout/check-in history      │   │
│  └──────────────────────────────────────────────────────────┘   │
│                    (Port 27017)                                 │
│  - Persistent volume for data durability                        │
│  - Atomic operations for consistency                            │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Docker Compose Environment                   │
│                                                                 │
│           ┌──────────────────┐    ┌──────────────────┐          │
│           │                  │    │                  │          │
│           │  Frontend Image  │    │  Backend Image   │          │
│           │  (Node.js Vite)  │    │  (Python 3.12)   │          │
│           │                  │    │                  │          │
│           └────────┬─────────┘    └────────┬─────────┘          │
│                    │                       │                    │
│             ┌──────▼─────────┬─────────────▼──────┐             │
│             │ Port 5173      │  Port 5001         │             │
│             │ React Dev      │  Flask API         │             │
│             └────────────────┴────────────────────┘             │
│                               │                                 │
│                       ┌───────▼──────────┐                      │
│                       │  Internal Network │                     │
│                       │  bridge           │                     │
│                       └───────┬──────────┘                      │
│                               │                                 │
│                       ┌───────▼──────────┐                      │
│                       │  Mongo Container │                      │
│                       │  Port 27017      │                      │
│                       │  Volume: mongo_data                     │
│                       └──────────────────┘                      │
│                                                                 │
│           Startup Scripts:                                      │
│           - scripts/start_app.sh (starts all services)          │
│           - scripts/start_backend.sh (backend only)             │
│           - scripts/start_frontend.sh (frontend only)           │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 Network Flow

1. **Frontend to Backend**: Axios via JavaScript/HTTP
   - Base URL: `/api` (proxied in dev by Vite to `http://localhost:5001`)
   - All requests include CORS headers
   - Request/response interceptors for logging and token management

2. **Backend to Database**: PyMongo via MongoDB wire protocol
   - Connection pooling managed by MongoClient
   - Atomic operations via MongoDB transactions (where applicable)

3. **Session Management**:
   - Login/registration handled via `/api/auth/login` and `/api/users` endpoints
   - Authentication state stored in React Context (client-side)
   - Backend validates each request independently (stateless)

---

## 5. Backend Design

### 3.1 Technology Stack

- **Framework**: Flask 3.1.2
- **Language**: Python 3.12+
- **Database Driver**: PyMongo 4.16.0
- **CORS**: Flask-CORS 6.0.2
- **Environment**: python-dotenv 1.2.1
- **Web Server**: Flask development server (gunicorn in production)

### 3.2 Project Structure

```
backend/
├── app/
│   ├── __init__.py          # App factory, blueprint registration
│   ├── config.py            # Configuration management
│   ├── db.py                # MongoDB connection & utilities
│   ├── mongo_utils.py       # Document serialization helpers
│   └── routes/
│       ├── __init__.py
│       ├── auth.py          # Authentication: /api/auth/login
│       ├── users.py         # Users: /api/users (CRUD)
│       ├── projects.py      # Projects: /api/projects (CRUD)
│       ├── hardware.py       # Hardware: /api/hardware* (TO BUILD)
│       ├── health.py        # Health: /api/health
│       └── root.py          # Root: /
├── run.py                   # Entry point
├── pyproject.toml           # Dependencies & metadata
├── Dockerfile               # Container configuration
└── .flaskenv               # Flask environment variables
```

### 3.3 Core Components

#### 3.3.1 Application Factory (app/__init__.py)

**Responsibility**: Initialize Flask app with all extensions and blueprints

**Current Status**: IMPLEMENTED

**Design Details**:
- Uses Flask application factory pattern
- Initializes MySQL connection via `init_mongo()`
- Registers all route blueprints with appropriate URL prefixes
- Configures CORS for all `/api/*` endpoints
- Sets up cleanup handlers for graceful shutdown

**Code Flow**:
```python
def create_app() -> Flask:
    # 1. Create Flask instance
    # 2. Load configuration from Config class
    # 3. Initialize CORS with frontend origins
    # 4. Initialize MongoDB connection
    # 5. Register route blueprints:
    #    - auth: /api/auth
    #    - users: /api/users
    #    - projects: /api/projects
    #    - health: /api/health (root: /)
    # 6. Register cleanup handler
    # 7. Return configured app
```

#### 3.3.2 Database Connection (app/db.py)

**Responsibility**: Manage MongoDB connection lifecycle and database access

**Current Status**: IMPLEMENTED

**Design Details**:
- Single MongoDB client per app instance
- Connection pooling configured by PyMongo
- `init_mongo()`: Establishes connection with 5-second timeout
- `get_db()`: Returns database instance from current app context
- `close_mongo()`: Cleanup on app shutdown

**Key Functions**:
- `init_mongo(app)`: Initialize connection, test with `ping` command
- `get_db() -> Database`: Get database for current request
- `close_mongo(app)`: Close connection on shutdown

**Configuration**:
```python
MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
MONGO_DB = os.getenv("MONGO_DB", "test")
```

#### 3.3.3 Authentication Module (app/routes/auth.py)

**Responsibility**: Handle user authentication and credential encryption

**Current Status**: PARTIALLY IMPLEMENTED

**Status Mapping**:
- `_encrypt()` function: IMPLEMENTED
- `/api/auth/login` endpoint: IMPLEMENTED
- `/api/auth/register` endpoint: NEEDS IMPLEMENTATION (use POST /api/users instead)

**Design Details**:

**Encryption Algorithm (F3/E1 Cipher)**:
- Current implementation: Simple cyclic cipher with shift
- Process:
  1. Validate input (ASCII only, no spaces/exclamation)
  2. Reverse input string
  3. Shift all characters by `num_shift * direction` positions
  4. Wrap around at ASCII boundaries
- Parameters:
  - `num_shift`: Number of positions to shift (typically 3)
  - `direction`: +1 (forward) or -1 (backward)

**Security Notes**:
- Current cipher is WEAK - suitable only for PoC
- FUTURE: Upgrade to bcrypt or Argon2
- Passwords stored ENCRYPTED in MongoDB
- Userid also encrypted for additional security

**Login Endpoint**:
```
POST /api/auth/login
Input: { userid, password }
Process:
  1. Encrypt both userid and password with same algorithm
  2. Query users collection for matching encrypted values
  3. Return user details (without password) on success
  4. Return 401 Unauthorized if no match
Output: { ok: true/false, message, user: { userid } }
```

#### 3.3.4 Users Module (app/routes/users.py)

**Responsibility**: User account management (create, read, update, delete)

**Current Status**: IMPLEMENTED

**Design Details**:

**Collection Schema**:
```
users: {
  _id: ObjectId (auto-generated),
  userId: String (unique, encrypted),
  password: String (encrypted),
  [other fields]: optional
}
```

**Note**: Per professor clarification, only `userId` and `password` are required fields (no username).

**Endpoints**:

| Method | Endpoint | Status | Details |
|--------|----------|--------|---------|
| POST | /api/users | IMPLEMENTED | Create user with validation |
| GET | /api/users | IMPLEMENTED | List up to 200 users |
| GET | /api/users/<id> | IMPLEMENTED | Get single user by MongoDB _id |
| PUT | /api/users/<id> | IMPLEMENTED | Replace user document |
| PATCH | /api/users/<id> | IMPLEMENTED | Partial user update |
| DELETE | /api/users/<id> | IMPLEMENTED | Delete user |

**Create User (POST /api/users)**:
```
Input: {
  userId: String (must be unique),
  password: String
}

Validation:
  - All required fields present (userId, password) - per professor clarification
  - `userId` MUST be unique (checked via MongoDB unique index)
  - Both userId and password must be encrypted
  - Fields are non-empty strings

Processing:
  1. Validate all required fields present
  2. Check userId uniqueness (prevent duplicates)
  3. Encrypt both userId and password using cyclic cipher (shift 3, direction 1)
  4. Insert into users collection
  5. Retrieve created document

Output: {
  _id: String,
  userId: String (decrypted for client display),
  password: String (NOT returned - redacted)
}

Error Codes:
  - 400: Missing/invalid fields
  - 409: userId already exists (duplicate key)
  - 500: Database error
```

**Security Feature**: `sanitize_user()` function removes sensitive fields (userid, password) from ALL API responses to prevent leaking the unique login identity.

#### 3.3.5 Projects Module (app/routes/projects.py)

**Responsibility**: Project management (create, read, filter, manage)

**Current Status**: IMPLEMENTED

**Design Details**:

**Collection Schema**:
```
projects: {
  _id: ObjectId,
  projectId: String (unique),
  name: String,
  description: String,
  ownerUserId: String,
  createdAt: Date (optional)
}
```

**Endpoints**:

| Method | Endpoint | Status | Details |
|--------|----------|--------|---------|
| POST | /api/projects | IMPLEMENTED | Create new project |
| GET | /api/projects | IMPLEMENTED | List projects (with optional filter) |
| GET | /api/projects/<id> | IMPLEMENTED | Get single project by MongoDB _id |
| PUT | /api/projects/<id> | IMPLEMENTED | Replace project |
| PATCH | /api/projects/<id> | IMPLEMENTED | Partial project update |
| DELETE | /api/projects/<id> | IMPLEMENTED | Delete project |

**Create Project (POST /api/projects)**:
```
Input: {
  projectId: String (unique identifier),
  name: String,
  description: String,
  ownerUserId: String (optional, should be current user)
}

Validation:
  - All required fields present and non-empty
  - projectId must be unique

Processing:
  1. Validate input
  2. Check projectId uniqueness
  3. Insert into projects collection
  4. Return created document

Error Codes:
  - 400: Missing/invalid fields
  - 409: projectId already exists
```

**Filter by Owner (GET /api/projects?ownerUserId=<userId>)**:
```
Query: ownerUserId (optional)
If provided:
  - Filter projects by matching ownerUserId
  - Supports user viewing their own projects

Returns: Array of project documents (limited to 200)
```

**Field Validation Rules**:
- projectId: non-empty, unique constraint enforced
- name: non-empty, human-readable
- description: non-empty, descriptive text
- ownerUserId: non-empty string matching user id

#### 3.3.6 Hardware Module (app/routes/hardware.py)

**Responsibility**: Hardware resource management and allocation tracking

**Current Status**: NOT YET IMPLEMENTED

**Design Details**:

This module MUST be created to handle hardware resource management. Based on the brain dump and SRS requirements, here's the design:

**Collection Schemas**:

**hardware_sets**:
```
{
  _id: ObjectId,
  hwSetId: String (unique: "HWSet1", "HWSet2", etc.),
  name: String,
  capacity: Integer (total available units),
  specifications: {
    cpu: String,
    ram: String,
    storage: String
  }
}
```

**resource_requests**:
```
{
  _id: ObjectId,
  projectId: String,
  userId: String,
  hwSet: String (reference to hwSetId),
  quantityRequested: Integer,
  status: String (enum: "pending", "approved", "checkout"),
  timestamp: ISODate
}
```

**allocations**:
```
{
  _id: ObjectId,
  projectId: String,
  userId: String,
  hwSet: String,
  units: Integer,
  type: String (enum: "checkout", "checkin"),
  timestamp: ISODate,
  status: String (enum: "active", "returned")
}
```

**Required Endpoints**:

| Method | Endpoint | Purpose | Status |
|--------|----------|---------|--------|
| GET | /api/hardware | List all hardware sets | NEED |
| POST | /api/hardware/request | Request hardware resources | NEED |
| POST | /api/hardware/checkout | Checkout resources | NEED |
| POST | /api/hardware/checkin | Check-in resources | NEED |
| GET | /api/hardware/availability | Get available units per set | NEED |
| GET | /api/hardware/allocations | View allocation history | NEED |

**Business Logic Requirements**:
1. **Capacity Management**: 
   - Calculate available = totalCapacity - allocatedUnits
   - Use MongoDB `$inc` operator for atomic availability updates
   - Prevent over-allocation at checkout time

2. **Availability Calculation**:
   - Real-time computation based on active allocations
   - Aggregate sum of allocated units by hardware set
   - Return in dashboard-friendly format

3. **Request Workflow**:
   - pending → approved → checkout → checkin
   - Validate requested quantity ≤ available
   - Prevent duplicate checkout without return

4. **Atomic Operations**:
   - Use MongoDB atomic operations for concurrent safety
   - Increment/decrement available units via $inc operator
   - MongoDB transactions for multi-step operations if needed

### 3.4 Error Handling Strategy

**Standard HTTP Status Codes**:
- `200 OK`: Successful request
- `201 Created`: Resource created (POST responses)
- `204 No Content`: Successful deletion
- `400 Bad Request`: Invalid input or validation failure
- `401 Unauthorized`: Authentication required/failed
- `404 Not Found`: Resource doesn't exist
- `409 Conflict`: Resource conflict (duplicate key, insufficient capacity)
- `500 Internal Server Error`: Server error

**Response Format**:
```json
// Success
{
  "ok": true,
  "message": "Operation successful",
  "data": { /* resource */ }
}

// Error
{
  "ok": false,
  "error": "Descriptive error message",
  "status": 400
}
```

### 3.5 Configuration Management

**File**: `app/config.py`

**Environment Variables**:
```python
MONGO_URI = "mongodb://localhost:27017"  # MongoDB connection
MONGO_DB = "test"                         # Database name
CORS_ORIGINS = ["http://localhost:5173"] # Frontend origins
FLASK_DEBUG = "1"                         # Debug mode (dev only)
```

**Loading**:
- Environment variables from `.env` file (via python-dotenv)
- Fallback to defaults if not specified
- CORS_ORIGINS can be comma-separated list

---

## 6. Frontend Design

### 4.1 Technology Stack

- **Framework**: React 19.2.0
- **Language**: TypeScript 5.9+
- **Build Tool**: Vite 7.2.4
- **Router**: React Router 7.13.0
- **HTTP Client**: Axios 1.13.4
- **UI Library**: Material-UI (MUI) 7.3.7
- **Build Output**: SPA (Single Page Application)

### 4.2 Project Structure

```
frontend/
├── src/
│   ├── main.tsx              # Application entry point
│   ├── App.tsx               # Root component
│   ├── App.css               # Global styles
│   ├── index.css             # Global CSS
│   │
│   ├── api/                  # API integration layer
│   │   ├── http.ts           # Axios instance with interceptors
│   │   ├── users.ts          # User-related API calls
│   │   ├── projects.ts       # Project-related API calls
│   │   └── hardware.ts       # Hardware API calls (TO BUILD)
│   │
│   ├── auth/                 # Authentication management
│   │   ├── authContext.ts    # React Context type definitions
│   │   ├── AuthProvider.tsx  # Context provider component
│   │   ├── useAuth.ts        # Custom hook for auth context
│   │   └── index.ts          # Barrel export
│   │
│   ├── pages/                # Page components (routes)
│   │   ├── Auth.tsx          # Login/Register page
│   │   ├── Home.tsx          # Home/Dashboard page
│   │   ├── Account.tsx       # User account page
│   │   ├── Projects.tsx      # Project management (TO BUILD)
│   │   ├── Hardware.tsx      # Hardware management (TO BUILD)
│   │   └── index.ts          # Barrel export
│   │
│   ├── layouts/              # Layout components
│   │   ├── AppLayout.tsx     # Main layout (header, footer, nav)
│   │   └── index.ts          # Barrel export
│   │
│   ├── routes/               # Routing configuration
│   │   ├── router.tsx        # Route definitions & structure
│   │   ├── ProtectedRoute.tsx# Protected route wrapper
│   │   └── index.ts          # Barrel export
│   │
│   └── assets/               # Static assets (images, icons, etc.)
│
├── public/                   # Public static files
├── index.html               # HTML template
├── vite.config.ts           # Vite configuration with API proxy
├── tsconfig.json            # TypeScript root configuration
├── tsconfig.app.json        # TypeScript app configuration
├── tsconfig.node.json       # TypeScript node configuration
├── eslint.config.js         # ESLint configuration
├── package.json             # Dependencies & scripts
└── README.md               # Project documentation
```

### 4.3 Core Components

#### 4.3.1 Entry Point & App Component (main.tsx, App.tsx)

**Responsibility**: Bootstrap React application and set up context providers

**Current Status**: IMPLEMENTED

**Design Details**:

**main.tsx**:
```typescript
// Creates React root and mounts App component
// Wrapped in StrictMode for development checks
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
```

**App.tsx**:
```typescript
// Wraps entire app with AuthProvider for context
// Provides RouterProvider for React Router
export default App() {
  return (
    <AuthProvider>
      <RouterProvider router={router} />
    </AuthProvider>
  )
}
```

#### 4.3.2 Authentication System

**Status**: PARTIALLY IMPLEMENTED

**Components**:

**AuthContext & AuthProvider** (auth/authContext.ts, auth/AuthProvider.tsx)

**Current Implementation Issues**:
- Using `email` field instead of `userId`
- Should be updated to use `userId` for consistency with backend

**Design**:
```typescript
type AuthUser = {
  _id: string;        // MongoDB object ID
  userId: string;     // User identifier (encrypted on backend)
};

type AuthContextValue = {
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (userId: string) => void;
  logout: () => void;
};
```

**Context Provider**:
- Maintains authentication state
- Provides login/logout functions
- Makes state available to all descendant components via useAuth() hook

**useAuth Hook** (auth/useAuth.ts):
```typescript
// Custom hook to access auth context
// Throws error if used outside AuthProvider
const { user, isAuthenticated, login, logout } = useAuth();
```

**NEEDS UPDATE**:
1. Rename `email` → `userId` in AuthUser type
2. Update AuthProvider to use userId instead of email
3. Update login function to accept userId parameter
4. Verify integration with login API response

#### 4.3.3 HTTP Client (api/http.ts)

**Responsibility**: Axios instance with interceptors for API requests

**Current Status**: IMPLEMENTED

**Design Details**:

**Base Configuration**:
```typescript
const api = axios.create({
  baseURL: '/api',        // Proxied to http://localhost:5001/api
  timeout: 10000,         // 10 second timeout
  headers: {
    'Content-Type': 'application/json'
  }
});
```

**Request Interceptor**:
- Adds authentication token from localStorage if available
- Logs requests in development mode with request details
- Handles request errors

**Response Interceptor**:
- Logs successful responses in development
- Handles error responses with specific logic:
  - 401: Clear token and suggest re-authentication
  - 403: Log forbidden access
  - 500+: Log server errors
- Provides user-friendly error messages via `getErrorMessage()` helper

**Helper Functions**:
```typescript
// Standard HTTP methods with typing
api.get<T>(url, config)
api.post<T>(url, data, config)
api.put<T>(url, data, config)
api.patch<T>(url, data, config)
api.delete<T>(url, config)

// Error extraction
getErrorMessage(error): string
```

#### 4.3.4 API Modules

**File Structure**:
- `api/users.ts`: User authentication and management
- `api/projects.ts`: Project operations
- `api/hardware.ts`: Hardware resource operations (TO BUILD)

**Design Pattern**: Each module exports a namespace object with API functions

**Users Module** (api/users.ts):

**Types**:
```typescript
type User = { _id: string; userId: string };
type LoginRequest = { userId: string; password: string };
type RegisterRequest = { userId: string; password: string };
type LoginResponse = { user: User; token?: string; message: string };
type RegisterResponse = { user: User; message: string };
```

**API Functions**:
```typescript
usersApi.login(credentials: LoginRequest)
  // POST /api/auth/login
  // Returns: { user, message }

usersApi.register(userData: RegisterRequest)
  // POST /api/auth/register (or POST /api/users)
  // Returns: { user, message }
```

**Projects Module** (api/projects.ts):

**Types**:
```typescript
type Project = {
  _id: string;
  projectId: string;
  name: string;
  description: string;
  ownerUserId?: string;
};
```

**API Functions**:
```typescript
projectsApi.list(ownerUserId?: string)
  // GET /api/projects?ownerUserId=<id> (optional filter)
  // Returns: Project[]

projectsApi.create(project: {
  projectId: string;
  name: string;
  description: string;
  ownerUserId?: string;
})
  // POST /api/projects
  // Returns: Project
```

**Hardware Module** (api/hardware.ts) - TO BUILD:

**Expected Types**:
```typescript
type HardwareSet = {
  _id: string;
  hwSetId: string;      // "HWSet1", "HWSet2"
  name: string;
  capacity: number;      // Total units
  specifications?: object;
};

type HardwareAvailability = {
  hwSetId: string;
  totalCapacity: number;
  allocated: number;    // Currently checked out
  available: number;    // Available for checkout
  utilizationPercent: number;
};

type ResourceRequest = {
  _id?: string;
  projectId: string;
  hwSet: string;
  quantityRequested: number;
  status: string;       // pending, approved, checkout, checkin
};

type Allocation = {
  _id?: string;
  projectId: string;
  hwSet: string;
  units: number;
  type: string;         // checkout, checkin
  timestamp: string;
  status: string;       // active, returned
};
```

**Expected API Functions**:
```typescript
hardwareApi.list()
  // GET /api/hardware
  // Returns: HardwareSet[]

hardwareApi.getAvailability()
  // GET /api/hardware/availability
  // Returns: HardwareAvailability[]

hardwareApi.requestResources(request: ResourceRequest)
  // POST /api/hardware/request
  // Returns: ResourceRequest with generated ID

hardwareApi.checkout(request: ResourceRequest)
  // POST /api/hardware/checkout
  // Returns: Allocation

hardwareApi.checkin(allocationId: string)
  // POST /api/hardware/checkin
  // Returns: Allocation with checkinTime

hardwareApi.getAllocations(projectId?: string)
  // GET /api/hardware/allocations?projectId=<id>
  // Returns: Allocation[]
```

#### 4.3.5 Page Components

**Status**: Auth and Home IMPLEMENTED; Projects and Hardware TO BUILD

**Auth Page** (pages/Auth.tsx)

**Purpose**: Login and registration interface

**Current Status**: IMPLEMENTED

**Features**:
- Dual-mode form (login / register toggle)
- Form validation:
  - Required fields check
  - Password match for registration
  - Minimum password length (6 chars)
- Error handling with Alert component
- Loading state during submission
- Auto-login after successful registration
- Redirect to specified page on successful login

**Form Fields** (per professor clarification - Feb 2026):
- userId: Unique login identity input
- password: Password input
- confirmPassword: Password confirmation (register mode only)

**Field Semantics**:
- `userId`: Unique login identity, used for authentication, displayed in account page
- No username field required (professor confirmed only userId and password needed)

**Integration**:
- Calls `usersApi.login()` for authentication (uses userId + password)
- Calls `usersApi.register()` for account creation (uses userId + password only)
- Updates auth context via `login()` function
- Navigates to `/account` or previous page on success

**Status**: IMPLEMENTED - Registration uses userId and password per professor clarification

**Home Page** (pages/Home.tsx)

**Purpose**: Dashboard and navigation hub

**Current Status**: IMPLEMENTED

**Features**:
- Conditional rendering based on authentication status
- Authenticated view: Quick actions, navigate to account/projects
- Unauthenticated view: Welcome message, Sign In button
- Material-UI Card layout

**Account Page** (pages/Account.tsx)

**Purpose**: User account information and management

**Current Status**: IMPLEMENTED

**Features**:
- Display current user profile info
- Show User ID and Account ID
- Protected route (requires authentication)

**Projects Page** (pages/Projects.tsx) - TO BUILD

**Purpose**: Project management interface

**Expected Features**:
- List user's projects with filter options
- Create new project form
- Project cards/table with metadata
- Join existing project by projectId
- Navigation to project details

**Hardware Page** (pages/Hardware.tsx) - TO BUILD

**Purpose**: Hardware resource management and checkout

**Expected Features**:
- Hardware inventory display (HWSet1, HWSet2)
- Real-time availability visualization
- Resource request form (quantity, hardware set, project selection)
- Active allocations display
- Checkout/check-in actions
- Allocation history table
- Dashboard widgets showing utilization

#### 4.3.6 Layout & Routing

**AppLayout Component** (layouts/AppLayout.tsx)

**Purpose**: Main application shell with header, footer, navigation

**Current Status**: IMPLEMENTED

**Structure**:
```
┌─────────────────────────────────┐
│       AppBar (Header)            │
│  Home | Account  [Title]  SignIn│
└─────────────────────────────────┘

        [Page Content]
       (via Outlet)

┌─────────────────────────────────┐
│    AppBar (Footer)               │
│    © Year Cloud Native Team      │
└─────────────────────────────────┘
```

**Features**:
- Fixed header with navigation buttons
- Center-positioned title
- Dynamic auth button (Sign In / Sign Out)
- Fixed footer with copyright
- Main content area with proper spacing
- Responsive design

**Router Configuration** (routes/router.tsx)

**Current Status**: PARTIALLY IMPLEMENTED

**Route Structure**:
```
/
├── / (Home)
├── /auth (Auth page)
└── (Protected)
    └── /account (Account)
```

**TO ADD**:
- `/projects` - Project management
- `/hardware` - Hardware management
- `/allocations` - Allocation history

**ProtectedRoute Component** (routes/ProtectedRoute.tsx)

**Purpose**: Wrapper for routes requiring authentication

**Current Status**: IMPLEMENTED

**Logic**:
```
If user is authenticated:
  - Render Outlet (child routes)
Else:
  - Redirect to /auth with return location
```

**Usage**:
```tsx
{
  element: <ProtectedRoute />,
  children: [
    { path: "account", element: <Account /> }
  ]
}
```

### 4.4 State Management Strategy

**Client-Side State Management**: React Context API

**AuthContext** (shared authentication state):
- Global auth state accessible via `useAuth()` hook
- Stores user info and authentication flag
- Methods: login(), logout()

**Local Component State**: useState
- Form state in Auth.tsx
- Loading/error states during API calls
- UI state (modals, filters, etc.)

**Strategy Rationale**:
- No need for Redux complexity for simple PoC
- Context API sufficient for application requirements
- LocalStorage for auth token persistence (future enhancement)

---

## 7. Data Model Design

### 5.1 MongoDB Collections

**INTEGRATION NOTES**: This section integrates the detailed database schema provided by project requirements (Feb 10, 2026). All fields, types, and indexes specified below reflect the project's design specifications. Todo markers [DESIGN CONFLICT] indicate design conflicts that require team discussion before implementation.

#### 5.1.0 Design Conflicts & Critical Decisions

**Issue 1: Projects Collection - Single Owner vs. Team-Based Access**

Status: [DESIGN CONFLICT] **NEEDS TEAM DECISION**

Conflict Summary:
- **Current Implementation**: Uses single `ownerUserId: String` field
- **User Requirements**: Specify `assignedUsers: String[]` array for team-based access
- **Root Cause**: Original implementation prioritized single-owner simplicity; requirements indicate multi-team collaboration needed

Decision Impact:
| Aspect | Single Owner (ownerUserId) | Team-Based (assignedUsers[]) |
|--------|---------------------------|------------------------------|
| Team Collaboration | No (not supported) | Yes (full support) |
| Access Control Logic | Simple (owner check) | Complex (array membership) |
| Filtering/Queries | `{ownerUserId: id}` | `{assignedUsers: id}` |
| Update Complexity | Add/remove one owner | Add/remove array members |
| API Endpoint Needed | `GET /api/projects?ownerUserId=<id>` | `GET /api/projects?assignedUser=<id>` |

**[DESIGN CONFLICT - TODO]**: Team consensus required on:
1. Do we support team-based projects → if YES, implement `assignedUsers` array
2. Do we maintain single ownership → if YES, keep `ownerUserId` field
3. Hybrid approach? (owner + assignedUsers both)

Recommendation: **Implement `assignedUsers[]` array** for maximum collaboration flexibility. Can be implemented as:
```javascript
// New document format
{
  projectId: "proj001",
  projectName: "Team ML Project",
  assignedUsers: ["user1", "user2", "user3"],  // All team members
  createdAt: ISODate(...)
}

// Query: Find projects for a specific user
db.projects.find({assignedUsers: "user123"})
```

---

#### 5.1.1 Users Collection

**Purpose**: Store user account information with encrypted credentials

**Schema** (per professor clarification - only userId and password needed):
```mongodb
{
  _id: ObjectId (MongoDB auto-generated),
  userId: String (unique index, encrypted per SR3),
  password: String (encrypted per SR3),
  createdAt: Date
}
```

**Indexes**:
- `_id`: Primary index (auto)
- `userId`: Unique index (uniqueness constraint)

**Field Semantics** (per professor clarification):
- `userId`: User's login identifier, stored encrypted per SR3, decrypted for display

**Validation**:
- userId: Non-empty, unique (enforced by MongoDB unique index), encrypted before storage per SR3
- password: Non-empty, encrypted using cyclic cipher (shift 3, direction 1) before storage per SR3

**Usage**:
- User registration: Insert new document with encrypted userId and encrypted password
- User authentication: Encrypt input credentials and compare with stored values
- User retrieval: Query by _id; return decrypted userId for display
- Return decrypted userId in API responses for UI display

**Security Notes**:
- BOTH userId and password are encrypted before storage per SR3
- password is NEVER returned in API responses
- userId is returned DECRYPTED for display purposes
- Client receives decrypted userId for display

#### 5.1.2 Projects Collection

**Purpose**: Store project definitions with ownership and team member tracking

**[DESIGN CONFLICT] - NEEDS TEAM DISCUSSION**

Current implementation uses `ownerUserId` for single owner tracking, but project requirements specify `assignedUsers[]` for team-based access. **[TODO]**: Team to decide:
- Option A: Keep single `ownerUserId` (current implementation) - simpler but limits collaboration
- Option B: Switch to `assignedUsers: String[]` - supports multiple team members but requires access control logic
- Impact: Affects filtering, access control, and project visibility logic

**Proposed Schema** (with `assignedUsers[]` for team access):
```mongodb
{
  _id: ObjectId,
  projectId: String (unique index),
  projectName: String,
  description: String,
  assignedUsers: [ String ] (array of user IDs with access),
  createdAt: Date
}
```

**Current Implementation** (single owner):
```mongodb
{
  _id: ObjectId,
  projectId: String (unique index),
  name: String,
  description: String,
  ownerUserId: String (single owner only),
  createdAt: Date
}
```

**Field Definitions**:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| _id | ObjectId | Yes | Auto-generated MongoDB ID |
| projectId | String | Yes | Unique identifier for project |
| projectName / name | String | Yes | Human-readable project name |
| description | String | Yes | Project description/purpose |
| assignedUsers / ownerUserId | String OR String[] | Yes | **CONFLICT**: Team member IDs (array) vs single owner (string) |
| createdAt | Date | No | Timestamp of project creation |

**Indexes**:
- `_id`: Primary index (auto)
- `projectId`: Unique constraint
- `ownerUserId`: Support for filtering by owner (if using single owner model)
- `assignedUsers`: Support for filtering by team member (if using array model)

**Validation**:
- projectId: Unique, non-empty, 1-100 characters
- projectName: Non-empty, descriptive text, 1-200 characters
- description: Non-empty, descriptive text, 1-1000 characters
- assignedUsers / ownerUserId: Non-empty, must reference valid users in users collection

**Usage** (current - single owner):
- Project creation: Insert with user's ID as ownerUserId
- Project retrieval: Query by _id or projectId
- Owner filtering: `db.projects.find({ownerUserId: "<userId>"})`
- Team view: Return projects where current user = ownerUserId

**Usage** (proposed - team-based):
- Project creation: Insert with user's ID in assignedUsers[] array
- Project retrieval: Query by _id or projectId
- Team member filtering: `db.projects.find({assignedUsers: "<userId>"})`
- Team view: Return projects where current user IN assignedUsers[]

#### 5.1.3 Hardware Sets Collection

**Purpose**: Define available hardware resource sets and their specifications

**Schema** (from user requirements):
```mongodb
{
  _id: ObjectId,
  hwSetId: String (unique index, e.g., "HWSet1", "HWSet2"),
  hardwareName: String,
  capacity: Integer (total available units),
  availability: Integer (currently available units, calculated from projectAllotments),
  projectAllotments: [
    {
      projectId: String,
      checkedOut: Integer (units checked out to this project)
    }
  ],
  createdAt: Date
}
```

**Field Definitions**:

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| _id | ObjectId | Yes | Auto-generated MongoDB ID |
| hwSetId | String | Yes | Unique hardware set ID (e.g., "HWSet1", "HWSet2") |
| hardwareName | String | Yes | Human-readable hardware name |
| capacity | Integer | Yes | Total available units in this set |
| availability | Integer | Yes | Currently available units (capacity - sum of checkedOut) |
| projectAllotments | Array[Object] | Yes | Per-project allocation tracking |
| projectAllotments.projectId | String | Yes | Reference to projects collection |
| projectAllotments.checkedOut | Integer | Yes | Units currently checked out to this project |
| createdAt | Date | No | Timestamp of hardware set creation |

**Design Notes**:
- `hwSetId`: Unique identifier for hardware set (e.g., "HWSet1", "HWSet2")
- `hardwareName`: Human-readable hardware name
- `capacity`: Total units in this hardware set (immutable)
- `availability`: Real-time available units (calculated as capacity - sum of projectAllotments[*].checkedOut)
- `projectAllotments`: Array tracking which projects have checked out units

**Indexes**:
- `_id`: Primary index (auto)
- `hwSetId`: Unique constraint
- `projectAllotments.projectId`: Support filtering allocations by project

**Validation**:
- hwSetId: Unique, non-empty, recommended pattern: "HWSet[0-9]+"
- hardwareName: Non-empty, descriptive
- capacity: Integer, > 0 (total units available)
- availability: Integer, 0 ≤ availability ≤ capacity
- projectAllotments: Array of objects, each with valid projectId and non-negative checkedOut count

**Usage Pattern**:
- availability = capacity - SUM(projectAllotments[*].checkedOut)
- Supports per-project allocation tracking
- Allows viewing which projects have resources allocated
- Pre-calculate availability on write for query efficiency

**Example Documents**:
```json
{
  "hwSetId": "HWSet1",
  "name": "GPU Cluster",
  "description": "High-performance GPU nodes",
  "totalCapacity": 100,
  "specifications": {
    "cpu": "Intel Xeon Platinum",
    "ram": "256GB",
    "gpu": "NVIDIA A40",
    "storage": "2TB NVMe"
  }
}

{
  "hwSetId": "HWSet2",
  "name": "CPU Server Pool",
  "description": "General-purpose CPU servers",
  "totalCapacity": 50,
  "specifications": {
    "cpu": "AMD EPYC",
    "ram": "128GB",
    "storage": "1TB SSD"
  }
}
```

#### 5.1.4 Resource Requests Collection (TO DESIGN)

**Purpose**: Track user requests for hardware resources

**Schema**:
```mongodb
{
  _id: ObjectId,
  projectId: String,
  userId: String,
  hwSet: String (reference to hwSetId in hardware_sets),
  quantityRequested: Integer,
  status: String (enum: "pending", "approved", "checkout", "checkin", "rejected"),
  requestDetails: String (optional),
  timestamp: ISODate,
  approvalTimestamp: ISODate (nullable),
  createdAt: Date
}
```

**Indexes**:
- `_id`: Primary
- `projectId`: Support filtering by project
- `userId`: Support filtering by user
- `status`: Support filtering by request status

**Validation**:
- projectId: Must reference existing project
- userId: Must be valid user identifier
- hwSet: Must reference valid hardware set
- quantityRequested: Positive integer ≤ available units at request time
- status: Valid enum value

**Usage**:
- Request creation: Insert with status "pending"
- Request approval: Update status to "approved"
- Request transition: Move through workflow states
- List user requests: `db.resource_requests.find({userId: "<id>"})`
- Availability check: Count active allocations

#### 5.1.5 Allocations Collection (TO DESIGN)

**Purpose**: Track hardware checkouts and check-ins (resource allocation history)

**Schema**:
```mongodb
{
  _id: ObjectId,
  projectId: String,
  userId: String,
  hwSet: String (reference to hwSetId),
  units: Integer (number of units checked out),
  type: String (enum: "checkout", "checkin"),
  timestamp: ISODate,
  status: String (enum: "active", "returned"),
  requestId: ObjectId (reference to resource_requests._id),
  createdAt: Date
}
```

**Indexes**:
- `_id`: Primary
- `projectId`: Support filtering by project
- `userId`: Support filtering by user
- `hwSet`: Support calculating availability
- `status`: Support querying active allocations
- `createdAt`: Support timeline analysis

**Validation**:
- projectId: Must reference existing project
- userId: Must be valid user identifier
- hwSet: Must reference valid hardware_sets.hwSetId
- units: Positive integer, count of allocation units
- type: Must be "checkout" or "checkin"
- timestamp: Valid ISO 8601 date format
- status: Must be "active" or "returned"

**Usage**:
- Record checkout: Create document with type="checkout", status="active"
- Record check-in: Create document with type="checkin", status="returned"
- Calculate availability: Sum active checkouts per hardware set
- View history: `db.allocations.find({projectId: "<id>"})`

**Availability Calculation Logic**:
```
For each hardware set:
  totalCapacity = db.hardware_sets.findOne({hwSetId}).totalCapacity
  
  allocatedUnits = db.allocations.aggregate([
    {
      $match: {
        hwSet: hwSetId,
        status: "active",
        type: "checkout"
      }
    },
    {
      $group: {
        _id: null,
        totalUnits: { $sum: "$units" }
      }
    }
  ])
  
  availableUnits = totalCapacity - allocatedUnits
  utilizationPercent = (allocatedUnits / totalCapacity) * 100
```

### 5.2 Collection Relationships

```
users (1) ──┬─→ (many) projects (owned by user)
            │
            └─→ (many) resource_requests (requested by user)
                └─→ referenced in allocations

projects (1) ──┬─→ (many) resource_requests
               │
               └─→ (many) allocations (assigned to project)

hardware_sets (1) ──┬─→ (many) resource_requests (requested hwset)
                    │
                    └─→ (many) allocations (for this hwset)

resource_requests (1) ──→ (1) allocations (when checked out)
```

### 5.3 Data Validation Rules

**Validation Location**: Backend API Layer (app/routes/*.py)

**User Validation** (per professor clarification - only userId and password needed):
- `userId`: String, required, unique, encrypted before storage per SR3
- `password`: String, required, non-empty, encrypted before storage per SR3

**Project Validation**:
- `projectId`: String, required, unique, 1-100 characters
- `name`: String, required, non-empty, 1-200 characters
- `description`: String, required, non-empty, 1-1000 characters
- `ownerUserId`: String, required, must match existing user

**Hardware Set Validation**:
- `hwSetId`: String, required, unique, matches pattern "HWSet[0-9]+"
- `name`: String, required, non-empty
- `totalCapacity`: Integer, required, > 0
- `specifications`: Object, optional, free-form

**Resource Request Validation**:
- `projectId`: String, required, must reference existing project
- `userId`: String, required, must be authenticated user
- `hwSet`: String, required, must reference hardware_sets.hwSetId
- `quantityRequested`: Integer, required, 1 ≤ qty ≤ available

**Allocation Validation**:
- `projectId`: String, required
- `userId`: String, required
- `hwSet`: String, required, valid hardware set reference
- `units`: Integer, required, > 0
- `type`: String, required, one of ["checkout", "checkin"]
- `timestamp`: Date, auto-set to current time
- `status`: String, enum: ["active", "returned"]

### 5.4 Capacity Management & Atomic Operations

**Capacity Calculation**:
```
For checkout validation:
  available = totalCapacity - SUM(active allocation units)
  
  if (quantityRequested > available) {
    return 409 Conflict (insufficient resources)
  }
```

**Atomic Update**: Use MongoDB's `$inc` operator
```javascript
// Atomic increment of allocated count
db.hardware_sets.updateOne(
  { hwSetId: "HWSet1" },
  { $inc: { allocatedUnits: requestedQuantity } }
)

// Atomic decrement on check-in
db.hardware_sets.updateOne(
  { hwSetId: "HWSet1" },
  { $inc: { allocatedUnits: -returnedQuantity } }
)
```

**Race Condition Prevention**:
1. Check availability before checkout
2. Use atomic increment operation
3. Insert allocation record after atomic update
4. Return 409 if capacity exhausted during operation

---

## 8. UML Class Diagrams

### 8.1 Data Model Class Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Domain Class Model                                 │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────┐
│       User           │
├──────────────────────┤
│ - _id: ObjectId      │
│ - userId: String     │ (encrypted per SR3)
│ - password: String   │ (encrypted per SR3)
│ - createdAt: Date    │
├──────────────────────┤
│ + register()         │
│ + login()            │
│ + getProfile()       │
└──────────────────────┘
        ▲
        │ creates (1 → many)
        │
┌──────────────────────┐          ┌──────────────────────────┐
│     Project          │◄─────────┤  User (ownerUserId)      │
├──────────────────────┤          └──────────────────────────┘
│ - _id: ObjectId      │
│ - projectId: String  │ (unique)
│ - projectName: Str   │
│ - description: Str   │
│ - ownerUserId: Str   │
│ - createdAt: Date    │
├──────────────────────┤
│ + create()           │
│ + getDetails()       │
│ + update()           │
│ + delete()           │
└──────────────────────┘
        ▲
        │ requests (1 → many)
        │
┌────────────────────────────────┐
│   ResourceRequest              │
├────────────────────────────────┤
│ - _id: ObjectId                │
│ - projectId: String (FK)       │
│ - userId: String (FK)          │
│ - hwSet: String (FK)           │
│ - quantityRequested: Integer   │
│ - status: Enum                 │
│ - timestamp: Date              │
├────────────────────────────────┤
│ + create()                     │
│ + approve()                    │
│ + reject()                     │
│ + getStatus()                  │
└────────────────────────────────┘
        ▲
        │ fulfills with (1 → many)
        │
┌──────────────────────────────┐
│    Allocation                │
├──────────────────────────────┤
│ - _id: ObjectId              │
│ - projectId: String (FK)     │
│ - userId: String (FK)        │
│ - hwSet: String (FK)         │
│ - units: Integer             │
│ - type: Enum (checkout/in)   │
│ - timestamp: Date            │
│ - status: Enum (active/ret)  │
├──────────────────────────────┤
│ + checkout()                 │
│ + checkin()                  │
│ + getHistory()               │
└──────────────────────────────┘

┌──────────────────────────────┐
│   HardwareSet                │
├──────────────────────────────┤
│ - _id: ObjectId              │
│ - hwSetId: String (unique)   │
│ - hardwareName: String       │
│ - capacity: Integer          │
│ - availability: Integer      │
│ - projectAllotments: Array   │
│ - specifications: Object     │
│ - createdAt: Date            │
├──────────────────────────────┤
│ + getAvailable()             │
│ + allocate()                 │
│ + deallocate()               │
│ + getUtilization()           │
└──────────────────────────────┘
        ▲
        │ managed by (1 → many Allocations)
        │
        └─ tracked in ResourceRequest
```

### 8.2 API Service Class Diagram

```
┌─────────────────────────────────────────────┐
│          API Service Layer                  │
└─────────────────────────────────────────────┘

Backend:
  ┌──────────────────────┐
  │   Flask Application  │
  ├──────────────────────┤
  │ + create_app()       │
  │ + initialize()       │
  │ + register_routes()  │
  └──────────────────────┘
            │
      ┌─────┼─────┐
      ▼     ▼     ▼
  ┌──────────────────────┬──────────────────────┬──────────────────────┐
  │  AuthRoute           │  UserRoute           │  ProjectRoute        │
  ├──────────────────────┼──────────────────────┼──────────────────────┤
  │ + login(userid, pwd) │ + create()           │ + create()           │
  │ + verify()           │ + getAll()           │ + getAll(filter)     │
  │ + encrypt()          │ + getById(id)        │ + getById(id)        │
  └──────────────────────┴──────────────────────┴──────────────────────┘
            │                   │                      │
            ▼                   ▼                      ▼
  ┌──────────────────────┬──────────────────────┬──────────────────────┐
  │  MongoDB Users Coll  │  MongoDB Users Coll  │  MongoDB Projects    │
  └──────────────────────┴──────────────────────┴──────────────────────┘

Frontend:
  ┌──────────────────────┐
  │   React Application  │
  ├──────────────────────┤
  │ - App.tsx            │
  │ - main.tsx           │
  └──────────────────────┘
            │
      ┌─────┼──────────┐
      ▼     ▼          ▼
  ┌──────────────────┬──────────────────┬──────────────────┐
  │  Axios Instance │  AuthProvider    │  Router Config   │
  ├──────────────────┼──────────────────┼──────────────────┤
  │ - baseURL        │ - context        │ - routes         │
  │ - interceptors   │ - login/logout   │ - protected      │
  └──────────────────┴──────────────────┴──────────────────┘
            │                │                │
            ▼                ▼                ▼
  ┌────────────────────────────────────────────────────────┐
  │              API Modules                               │
  ├────────────────────────────────────────────────────────┤
  │ - usersApi (login, register)                           │
  │ - projectsApi (list, create, update)                   │
  │ - hardwareApi (list, request, checkout, checkin)       │
  └────────────────────────────────────────────────────────┘
```

---

## 9. API Design

### 9.1 REST API Overview

**Base URL**: `http://localhost:5001/api`

**Response Format**: All responses are JSON

**Standard Response Wrapper**:
```json
{
  "ok": true,
  "message": "Success message",
  "data": { /* response data */ }
}
```

**Error Response**:
```json
{
  "ok": false,
  "error": "Error message",
  "status": 400
}
```

### 9.2 API Endpoints

#### 9.2.1 Authentication Endpoints

**POST /api/auth/login** - User Login

**Status**: IMPLEMENTED

```
Method: POST
URL: /api/auth/login
Request: {
  "userid": String,
  "password": String
}

Response 200: {
  "ok": true,
  "message": "Login successful",
  "user": {
    "userid": String (unencrypted for client use)
  }
}

Response 401: {
  "ok": false,
  "error": "Invalid credentials"
}
```

#### 9.2.2 User Management Endpoints

**POST /api/users** - Create User (Register)

**Status**: IMPLEMENTED

```
Method: POST
URL: /api/users/register
Request: {
  "userId": String,
  "password": String
}

Response 201: {
  "_id": String,
  "userId": String (decrypted for display)
}

Response 400: {
  "error": "Missing fields: [...]", 
   status: 400
}

Response 409: {
  "error": "User ID already exists",
  "status": 409
}
```

**GET /api/users** - List Users

**Status**: IMPLEMENTED

```
Method: GET
URL: /api/users

Response 200: [
  {
    "_id": String,
    "userId": String (decrypted for display)
  },
  ...
] (max 200 users)
```

**GET /api/users/<id>** - Get User Details

**Status**: IMPLEMENTED

```
Method: GET
URL: /api/users/<mongodb_id>

Response 200: {
  "_id": String,
  "userId": String (decrypted for display)
}

Response 404: {
  "error": "Not found"
}
```

#### 9.2.3 Project Management Endpoints

**POST /api/projects** - Create Project

**Status**: IMPLEMENTED

```
Method: POST
URL: /api/projects
Request: {
  "projectId": String (unique),
  "name": String,
  "description": String,
  "ownerUserId": String (optional)
}

Response 201: {
  "_id": String,
  "projectId": String,
  "name": String,
  "description": String,
  "ownerUserId": String
}

Response 400: Missing/invalid fields
Response 409: projectId already exists
```

**GET /api/projects** - List Projects (with optional filter)

**Status**: IMPLEMENTED

```
Method: GET
URL: /api/projects?ownerUserId=<userId>

Query Parameters:
  - ownerUserId: (optional) Filter by project owner

Response 200: [
  {
    "_id": String,
    "projectId": String,
    "name": String,
    "description": String,
    "ownerUserId": String
  },
  ...
] (max 200 projects)
```

**GET /api/projects/<id>** - Get Project Details

**Status**: IMPLEMENTED

#### 9.2.4 Hardware Management Endpoints (TO BUILD)

**GET /api/hardware** - List Hardware Sets

```
Method: GET
URL: /api/hardware

Response 200: [
  {
    "_id": String,
    "hwSetId": String,
    "name": String,
    "description": String,
    "totalCapacity": Integer,
    "specifications": { /* hardware specs */ }
  },
  ...
]
```

**GET /api/hardware/availability** - Get Real-Time Availability

```
Method: GET
URL: /api/hardware/availability

Response 200: [
  {
    "hwSetId": String,
    "totalCapacity": Integer,
    "allocatedUnits": Integer,
    "availableUnits": Integer,
    "utilizationPercent": Double
  },
  ...
]
```

**POST /api/hardware/request** - Request Hardware Resources

```
Method: POST
URL: /api/hardware/request
Request: {
  "projectId": String,
  "hwSet": String,
  "quantityRequested": Integer,
  "requestDetails": String (optional)
}

Response 201: {
  "_id": String,
  "projectId": String,
  "hwSet": String,
  "quantityRequested": Integer,
  "status": "pending",
  "timestamp": ISODate
}

Response 400: Invalid input
Response 409: Insufficient capacity
Response 404: Project/HWSet not found
```

**POST /api/hardware/checkout** - Checkout Hardware

```
Method: POST
URL: /api/hardware/checkout
Request: {
  "requestId": String,
  "projectId": String,
  "units": Integer (optional, default from request)
}

Response 200: {
  "_id": String,
  "projectId": String,
  "hwSet": String,
  "units": Integer,
  "type": "checkout",
  "timestamp": ISODate,
  "status": "active"
}

Response 400: Invalid request
Response 409: Request not approved OR resources unavailable
```

**POST /api/hardware/checkin** - Check In Hardware

```
Method: POST
URL: /api/hardware/checkin
Request: {
  "allocationId": String
}

Response 200: {
  "_id": String,
  "type": "checkin",
  "timestamp": ISODate,
  "status": "returned",
  "units": Integer
}

Response 404: Allocation not found
Response 403: User not authorized
```

**GET /api/hardware/allocations** - View Allocation History

```
Method: GET
URL: /api/hardware/allocations?projectId=<id>&userId=<id>

Query Parameters:
  - projectId: (optional) Filter by project
  - userId: (optional) Filter by user

Response 200: [
  {
    "_id": String,
    "projectId": String,
    "userId": String,
    "hwSet": String,
    "units": Integer,
    "type": String,
    "timestamp": ISODate,
    "status": String
  },
  ...
]
```

### 9.3 HTTP Status Codes

- `200 OK`: Successful GET, POST, PUT, PATCH
- `201 Created`: Successful resource creation
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Invalid input, validation failure
- `401 Unauthorized`: Auth required or failed
- `403 Forbidden`: User not authorized for resource
- `404 Not Found`: Resource doesn't exist
- `409 Conflict`: Resource conflict, duplicate key, insufficient capacity
- `500 Internal Server Error`: Server error

### 9.4 CORS Configuration

**Frontend Origin:** `http://localhost:5173`

**Environment Variable:**
```
CORS_ORIGINS=http://localhost:5173
```

**Applied to:** All `/api/*` endpoints

**Production Consideration**: Update CORS_ORIGINS for production domains

---

## 10. Security Design

### 10.1 Authentication Architecture

**Current Implementation**: F3/E1 Cipher (PoC Only)

**Security Characteristics**:
- **Strength**: WEAK - Suitable only for proof of concept
- **Type**: Symmetric cyclic cipher with character shift
- **Purpose**: Encrypt userid and password before storage
- **Risk Level**: HIGH - DO NOT use in production

**Encryption Process**:
```
Input: plaintext string (e.g., "user123")
       num_shift: 3
       direction: 1 (forward)

Step 1: Reverse string       → "321resu"
Step 2: Shift each character → Apply shift formula
Step 3: Character wrapping   → Wrap at ASCII boundaries

Formula: encrypted_char = (char_code + (num_shift * direction)) mod 256

Output: Encrypted string stored in MongoDB
```

**Decryption Process**:
```
Input: Encrypted string from MongoDB
       num_shift: 3
       direction: -1 (backward)

Process: Reverse the encryption (shift in opposite direction)

Output: Original plaintext for comparison
```

**Security Issues with Current Approach**:
1. Cipher is reversible without secret key (known shift value)
2. No salt - same plaintext always produces same ciphertext
3. Vulnerable to frequency analysis
4. Character shift is deterministic and simple
5. ASCII wrapping creates predictable patterns

**CRITICAL UPGRADE REQUIREMENT**:
- [ ] TODO: Replace F3/E1 with bcrypt or Argon2
- [ ] TODO: Add password salt generation
- [ ] TODO: Implement key derivation function (PBKDF2)
- [ ] Estimated timeline: Before production deployment

### 10.2 Authorization & Access Control

**Backend Authorization Model**:
- Stateless (no session tokens)
- Each request validated independently
- User identity extracted from request context
- Project access controlled via ownerUserId matching

**Protected Resources**:
- User profiles: Only accessible to self (future: implement)
- Projects: Only accessible to owner (ownerUserId match)
- Hardware allocations: Based on project ownership
- Admin endpoints: Reserved for future implementation

**Access Control Rules**:
```
User Authentication:
  - Login credentials encrypted with F3/E1 cipher  
  - Encrypted values stored in MongoDB
  - Comparison done server-side (client never sees encryption key)
  
Resource Authorization:
  - User can only view/modify resources they own
  - Project owner = user who created project
  - Hardware allocations tied to project ownership
  - No cross-project visibility
```

### 10.3 Data Protection

**At-Rest Encryption**:
- Field-level encryption: userid, password encrypted before storage
- Database-level: MongoDB persistence (unencrypted on disk in dev)
- Production: Should enable MongoDB encryption at rest

**In-Transit Encryption**:
- HTTP/HTTPS: Currently HTTP in development
- Production: ENFORCE HTTPS for all API endpoints
- CORS validation: Verified origin headers
- No sensitive data in URLs (all sensitive data in request body)

**Data Masking**:
- `sanitize_user()` function removes sensitive fields from responses
- Password NEVER returned in API responses
- UserId returned DECRYPTED for UI display (per professor clarification)
- Sensitive fields only used server-side for authentication

### 10.4 Network Security

**CORS Configuration**:
```python
# Current Development
CORS_ORIGINS = ["http://localhost:5173"]

# Production (TODO)
CORS_ORIGINS = [
  "https://example.com",
  "https://www.example.com"
]
```

**API Endpoint Protection**:
- All `/api/*` endpoints protected by CORS headers
- CORS allows credentials in frontend requests (future)
- Invalid origins rejected by Flask-CORS middleware

**Frontend Security Practices**:
- No hardcoded API keys in code
- API keys stored in environment variables
- Frontend doesn't validate tokens (backend authoritative)
- React Context used for client-side auth state only

### 10.5 Database Security

**MongoDB Security Practices**:
- Local development: No authentication required
- Production: ENFORCE MongoDB authentication
- User isolation: Separate read/write permissions per service
- Query injection: Use PyMongo parameterized queries (no string concatenation)

**Index Security**:
- Unique indexes enforce userid/projectId uniqueness
- No sensitive data indexed for searchability
- Indexes optimized for common queries, not secret exposure

**Backup & Recovery**:
- MongoDB persistence volume for data durability
- Regular backup schedule (TODO: implement)
- Encrypted backup storage (TODO: implement)
- Tested recovery procedures (TODO: document)

### 10.6 Input Validation & Sanitization

**Backend Validation**:
```
All API endpoints perform strict validation:
- Required field presence check
- Type verification (String, Integer, etc.)
- Length constraints (min/max characters)
- Format validation (ObjectId, email, etc.)
- No SQL injection (using PyMongo drivers)
```

**Frontend Validation**:
- TypeScript prevents type-related bugs
- Form validation before API calls
- User feedback on validation errors
- No sensitive data logged to console in production

### 10.7 Audit & Logging

**Current Logging**:
- Request interceptors log API calls (development mode)
- Error responses include descriptive messages
- No sensitive data (passwords, userid) logged

**Production Logging Recommendations**:
- [ ] Centralized logging (ELK stack or Azure Application Insights)
- [ ] Audit trail for resource allocations
- [ ] User activity logging
- [ ] Failed authentication attempt tracking
- [ ] Encrypted log storage

### 10.8 Security Checklist

- [ ] Authentication upgraded from F3/E1 to bcrypt/Argon2
- [ ] HTTPS enforced in production
- [ ] MongoDB authentication enabled
- [ ] Database encryption at rest enabled
- [ ] Backup encryption enabled
- [ ] API rate limiting implemented
- [ ] CSRF tokens implemented (if session-based)
- [ ] Security headers configured (X-Frame-Options, etc.)
- [ ] Dependency vulnerability scanning enabled
- [ ] Penetration testing performed
- [ ] Security training for team
- [ ] Incident response plan documented

---

## 11. Business Rules

### 11.1 User Management Rules
1. UserId MUST be unique across all users
2. UserId and password MUST be encrypted before storage per SR3
3. Same encryption algorithm (cyclic cipher, shift 3, direction 1) MUST be used for both fields
4. Password MUST NOT be returned in API responses
5. UserId is returned DECRYPTED for display purposes
6. User session state MUST be managed client-side only (stateless backend)

### 11.2 Project Management Rules
1. ProjectId MUST be unique across all projects
2. Only project owner CAN view/modify project details
3. Users CAN join projects by providing valid projectId
4. Project metadata (name, description, owner) is immutable after creation (recommendation)

### 11.3 Hardware Resource Rules
1. Hardware Sets (HWSet1, HWSet2) have fixed total capacity
2. Available units = Total Capacity - Allocated Units
3. Real-time availability MUST be calculated (not cached)
4. Cannot checkout more units than currently available
5. Over-allocation MUST be prevented at checkout time

### 11.4 Resource Workflow Rules
1. Request → Pending → Approved → Checkout → Check-in
2. Status transitions are sequential (cannot skip states)
3. Only approved requests CAN be checked out
4. Only active (checked-out) allocations CAN be checked in
5. Cannot checkout same request twice without returning it first

### 11.5 Capacity Management Rules
1. Use atomic MongoDB operations for availability updates
2. Increment allocated units on successful checkout
3. Decrement allocated units on successful check-in
4. Prevent race conditions via atomic operations
5. Validate availability before allowing checkout

### 11.6 Data Persistence Rules
1. All user data MUST persist in MongoDB
2. Encrypted credentials MUST remain encrypted in storage
3. Allocation history MUST be retained for audit trail
4. Deleted resources MUST be soft-deleted or archived (future consideration)
5. Timestamps (ISO 8601) MUST record all state changes

### 11.7 Error Handling Rules
1. All errors MUST return appropriate HTTP status codes
2. Error responses MUST include descriptive messages
3. Validation failures MUST return 400 Bad Request
4. Authentication failures MUST return 401 Unauthorized
5. Resource conflicts MUST return 409 Conflict

---

## 12. Design Patterns & Workflows

### 12.1 Authentication Workflow Flowchart

```
┌──────────────────────────────────────────────────┐
│         User Authentication Flow                 │
└──────────────────────────────────────────────────┘

Start: User enters credentials (userid, password)
   │
   ▼
┌─────────────────────────────────────────┐
│ Frontend: Collect userid and password   │
│ Validate: Non-empty fields              │
└─────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────┐
│ POST /api/auth/login with credentials   │
│ Axios sends JSON payload                │
└─────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────┐
│ Backend: Receive login request          │
│ Extract userid and password             │
└─────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────┐
│ Backend: Encrypt userid and password    │
│ Using F3/E1 cipher (num_shift=3)        │
└─────────────────────────────────────────┘
   │
   ▼
┌─────────────────────────────────────────┐
│ Backend: Query MongoDB users collection │
│ Match: {userid: encrypted, password: .. │
└─────────────────────────────────────────┘
   │
   ├── No match found ──────────┐
   │                            ▼
   │                   Return 401 Unauthorized
   │                   Message: "Invalid credentials"
   │                            │
   │                            ▼
   │                   Frontend: Show error message
   │                   User can retry
   │
   └── Match found ────────────┐
                                ▼
                    Backend: Extract user document
                    Remove sensitive fields
                    Return 200 OK with user data
                                │
                                ▼
                    Frontend: Receive user response
                    Store user in AuthContext
                    Update isAuthenticated = true
                                │
                                ▼
                    Frontend: Navigate to /account
                    or previous page
                                │
                                ▼
                              END
```

### 12.2 User Registration Workflow

```
┌──────────────────────────────────────────────────┐
│         User Registration Flow                   │
└──────────────────────────────────────────────────┘

Start: User fills registration form
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: Collect data                     │
│  - userId (login identifier)               │
│  - password                                │
│  - confirmPassword                         │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: Validate                         │
│  - All fields non-empty                    │
│  - Password == confirmPassword             │
│  - Password length >= 6 chars              │
└────────────────────────────────────────────┘
   │
   ├── Validation fails ────────────────────┐
   │                                        ▼
   │                          Show validation error
   │
   └── Validation passes ───────────────────┐
                                            ▼
                        POST /api/users/register with data
                        Payload: {userId, password}
                                            │
                                            ▼
                        Backend: Receive registration
                        Validate all required fields
                        Check userId uniqueness
                                            │
                        ├── Duplicate userId ────┐
                        │                        ▼
                        │         Return 409 Conflict
                        │
                        └── Unique userId ──────┐
                                                ▼
                            Encrypt userid and password
                            Create user document
                            Insert into MongoDB
                                                │
                                                ▼
                            Return 201 Created
                            User data in response
                                                │
                                                ▼
                            Frontend: Store user in context
                            Auto-login (call login())
                            Navigate to /account
                                                │
                                                ▼
                                              END
```

### 12.3 Hardware Resource Request Workflow

```
┌──────────────────────────────────────────────────┐
│     Hardware Resource Request Flow               │
└──────────────────────────────────────────────────┘

Start: User needs hardware resources
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: Display available hardware sets  │
│ Call GET /api/hardware/availability        │
│ Show: capacity, allocation, available      │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ User: Select hardware set & quantity       │
│ Form: {hwSet, quantityRequested, project}  │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: Validate user input              │
│  - Quantity > 0                            │
│  - Quantity <= available units             │
│  - Hardware set selected                   │
│  - Project selected                        │
└────────────────────────────────────────────┘
   │
   ├── Validation fails ────────────────────┐
   │                                        ▼
   │                          Show error message
   │
   └── Validation passes ───────────────────┐
                                            ▼
                        POST /api/hardware/request
                        Backend: Verify request
                        Calculate available units
                                            │
                        ├── Insufficient units ...┐
                        │                        ▼
                        │        Return 409 Conflict
                        │
                        └── Sufficient units ───┐
                                                ▼
                            Create ResourceRequest
                            status = "pending"
                            Store in DB
                            Return 201 Created
                                                │
                                                ▼
                            Frontend: Show success
                            Request now in "pending" state
                            Admin approval needed
                                                │
                                                ▼
                            Admin reviews request
                            Updates status → "approved"
                                                │
                                                ▼
                            User: Initiates checkout
                            POST /api/hardware/checkout
                                                │
                                                ▼
                            Backend: Verify availability
                            Atomic operation: $inc allocation
                            Create Allocation record
                                                │
                                                ▼
                            Frontend: Show active allocation
                            Display check-in option
                                                │
                                                ▼
                            User: When done, check-in
                            POST /api/hardware/checkin
                                                │
                                                ▼
                            Backend: Decrement allocation
                            Record check-in time
                            Update status → "returned"
                                                │
                                                ▼
                            Allocation history updated
                            Resources released for reuse
                                                │
                                                ▼
                                              END
```

### 12.4 Project Management Workflow

```
┌──────────────────────────────────────────────────┐
│     Project Creation & Management Flow           │
└──────────────────────────────────────────────────┘

Start: Authenticated user creates project
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: Display project creation form    │
│ Fields: {projectId, name, description}    │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ User: Enter project details                │
│ Validate: All fields required, valid chars │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ Frontend: POST /api/projects               │
│ Payload: {projectId, name, description,    │
│           ownerUserId: currentUser._id}    │
└────────────────────────────────────────────┘
   │
   ▼
┌────────────────────────────────────────────┐
│ Backend: Validate project data             │
│  - All required fields present             │
│  - projectId non-empty                     │
│  - Check projectId uniqueness              │
└────────────────────────────────────────────┘
   │
   ├── Duplicate projectId ────────────────┐
   │                                       ▼
   │                   Return 409 Conflict
   │
   └── Valid project ──────────────────────┐
                                           ▼
                       Insert into projects collection
                       Set ownerUserId = current user
                       Return 201 Created
                                           │
                                           ▼
                       Frontend: Show success
                       Navigate to projects list
                       Project now visible to owner
                                           │
                                           ▼
                       User: View projects
                       GET /api/projects?ownerUserId=<id>
                                           │
                                           ▼
                       Backend: Return projects
                       Filter by ownerUserId
                       Return up to 200 projects
                                           │
                                           ▼
                       Frontend: Display projects
                       Show project cards/table
                       Options: View, Edit, Delete
                                           │
                                           ▼
                       User: Select project
                       View allocation history
                       Request new equipment
                                           │
                                           ▼
                                         END
```

### 12.5 Design Patterns Used

**Pattern 1: Model-View-Controller (MVC)**
- Model: MongoDB collections (users, projects, hardware_sets, etc.)
- View: React components (pages, layouts)
- Controller: Flask routes (blueprints) and API endpoints

**Pattern 2: Factory Pattern**
- Application factory in Flask (create_app())
- Initializes app once, reusable instances

**Pattern 3: Blueprint Pattern (Modular Routes)**
- Auth routes registered as blueprint
- Users routes registered as blueprint
- Projects routes registered as blueprint
- Hardware routes (to be built) as blueprint

**Pattern 4: Data Access Object (DAO)**
- Database access abstracted in routes
- PyMongo queries encapsulated
- Reusable query results

**Pattern 5: Provider Pattern (React)**
- AuthProvider wraps application
- Provides context to all descendant components
- useAuth custom hook accesses provider

**Pattern 6: Protected Route Pattern**
- ProtectedRoute wrapper component
- Verifies authentication before rendering
- Redirects to /auth if unauthenticated

**Pattern 7: Interceptor Pattern**
- Axios interceptors for request/response handling
- Automatic error handling
- Request/response logging

**Pattern 8: Atomic Operations (Database)**
- MongoDB $inc operator for atomic updates
- Prevents race conditions on availability updates
- Ensures data consistency

---

## 13. Testing Strategy

### 13.1 Testing Levels

**Unit Testing**:
- Test individual functions in isolation
- Backend: Test encryption, validation functions
- Frontend: Test component rendering, prop handling
- Expected coverage: 70-80%

**Integration Testing**:
- Test component interaction (frontend)
- Test API endpoint flow (backend)
- Mock external dependencies
- Expected coverage: Backend routes, authentication flow

**End-to-End Testing**:
- Test complete user workflows
- Full-stack: frontend → backend → database
- Scenarios:
  - User registration and login
  - Project creation and management
  - Hardware resource request and checkout
  - Resource allocation tracking

**Performance Testing**:
- Load testing: Concurrent API requests
- Database query optimization
- Frontend rendering performance
- Target: Sub-100ms API response times

### 13.2 Backend Testing Plan

**Test Framework**: pytest

**Test Categories**:

1. **Authentication Tests** (`test_auth.py`)
   ```python
   - test_login_success()
   - test_login_invalid_credentials()
   - test_encrypt_decrypt()
   - test_password_encryption_consistency()
   ```

2. **User Management Tests** (`test_users.py`)
   ```python
   - test_create_user_success()
   - test_create_user_duplicate_userid()
   - test_create_user_missing_fields()
   - test_list_users()
   - test_get_user_by_id()
   - test_update_user()
   - test_delete_user()
   - test_user_sanitization()
   ```

3. **Project Management Tests** (`test_projects.py`)
   ```python
   - test_create_project_success()
   - test_create_project_duplicate_id()
   - test_list_projects_with_filter()
   - test_get_project_by_id()
   - test_update_project()
   - test_delete_project()
   - test_owner_authorization()
   ```

4. **Hardware Module Tests** (`test_hardware.py`)
   ```python
   - test_list_hardware_sets()
   - test_get_availability()
   - test_request_resources_success()
   - test_request_insufficient_capacity()
   - test_checkout_resources()
   - test_checkin_resources()
   - test_allocation_history()
   ```

5. **Database Tests** (`test_db.py`)
   ```python
   - test_connection_success()
   - test_connection_failure_handling()
   - test_document_insertion()
   - test_unique_indexes()
   - test_atomic_operations()
   ```

6. **Error Handling Tests** (`test_errors.py`)
   ```python
   - test_400_validation_error()
   - test_401_unauthorized()
   - test_404_not_found()
   - test_409_conflict()
   - test_500_server_error()
   ```

### 13.3 Frontend Testing Plan

**Test Framework**: Vitest + React Testing Library

**Test Categories**:

1. **Authentication Component Tests** (`Auth.test.tsx`)
   ```
   - test_login_form_submission()
   - test_register_form_submission()
   - test_form_validation_errors()
   - test_password_match_validation()
   - test_successful_login_navigation()
   ```

2. **Context Provider Tests** (`AuthProvider.test.tsx`)
   ```
   - test_login_function()
   - test_logout_function()
   - test_context_consumer_hook()
   - test_user_state_persistence()
   ```

3. **Protected Route Tests** (`ProtectedRoute.test.tsx`)
   ```
   - test_authenticated_access()
   - test_unauthenticated_redirect()
   - test_redirect_preservation()
   ```

4. **Page Component Tests** (`pages/*.test.tsx`)
   ```
   - test_home_page_rendering()
   - test_account_page_data_display()
   - test_projects_page_list()
   - test_hardware_page_availability()
   ```

5. **API Module Tests** (`api/*.test.ts`)
   ```
   - test_login_api_call()
   - test_register_api_call()
   - test_project_list_api_call()
   - test_api_error_handling()
   ```

6. **Layout Tests** (`layouts/AppLayout.test.tsx`)
   ```
   - test_header_rendering()
   - test_footer_rendering()
   - test_navigation_links()
   - test_responsive_layout()
   ```

### 13.4 Test Data & Fixtures

**Fixtures for Backend Testing**:
```python
# fixtures/users.py
test_user = {
  "userId": "test123",
  "password": "testpass"
}

# fixtures/projects.py
test_project = {
  "projectId": "proj001",
  "projectName": "Test Project",
  "description": "Test project for testing",
  "ownerUserId": "user123"
}

# fixtures/hardware.py
test_hardware_set = {
  "hwSetId": "HWSet1",
  "hardwareName": "GPU Cluster",
  "capacity": 100,
  "specifications": {...}
}
```

### 13.5 Continuous Integration (CI)

**GitHub Actions Workflow** (`.github/workflows/test.yml`):

```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.12'
      - name: Install dependencies
        run: pip install -e backend/
      - name: Run pytest
        run: pytest backend/tests/ -v --cov
      - name: Upload coverage
        run: codecov

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '20'
      - name: Install dependencies
        run: cd frontend && npm install
      - name: Run vitest
        run: cd frontend && npm run test
      - name: Build frontend
        run: cd frontend && npm run build
```

### 13.6 Manual Testing Checklist

Before each release:
- [ ] User registration happy path
- [ ] User login with valid credentials
- [ ] Login fails with invalid credentials
- [ ] Create project successfully
- [ ] Filter projects by owner
- [ ] View project details
- [ ] List available hardware sets
- [ ] Check hardware availability updates
- [ ] Request hardware resources
- [ ] Checkout resources successfully
- [ ] Check-in resources and verify return
- [ ] View allocation history
- [ ] Verify error messages display correctly
- [ ] Test navigation between all pages
- [ ] Verify responsive layout on mobile
- [ ] Test logout functionality

### 13.7 Known Issues & Limitations

**Current Testing Gaps**:
- [ ] No integration tests between frontend and backend
- [ ] No database fixtures/seeds
- [ ] No mock MongoDB server
- [ ] No performance benchmarks
- [ ] No security vulnerability scanning
- [ ] No end-to-end tests with Cypress/Playwright

**TODO: Implement Before Production**:
- [ ] Complete test coverage to 80%+
- [ ] Add E2E tests for all workflows
- [ ] Add database migration tests
- [ ] Add API performance tests
- [ ] Add security penetration tests
- [ ] Add load testing with k6 or JMeter

---

## 14. Implementation Status

### 14.1 Backend Status

| Component | Status | Notes |
|-----------|--------|-------|
| Flask App Setup | IMPLEMENTED | Application factory, blueprint registration |
| MongoDB Connection | IMPLEMENTED | Connection pooling, error handling |
| Configuration | IMPLEMENTED | Environment variable support |
| Authentication (/api/auth/login) | IMPLEMENTED | F3/E1 cipher encryption |
| Users Module (/api/users/*) | IMPLEMENTED | Full CRUD operations |
| Projects Module (/api/projects/*) | IMPLEMENTED | Full CRUD with owner filtering |
| Health Check (/api/health) | IMPLEMENTED | Basic status endpoint |
| Hardware Module (/api/hardware/*) | NEEDS BUILD | Resource management endpoints |
| Error Handling | NEEDS ENHANCEMENT | Should standardize response format |
| Input Validation | NEEDS ENHANCEMENT | Should add more type checking |
| Database Indexes | NEEDS BUILD | Should add indexes on foreign keys |

### 14.2 Frontend Status

| Component | Status | Notes |
|-----------|--------|-------|
| Project Setup | IMPLEMENTED | Vite, React, TypeScript configured |
| Axios HTTP Client | IMPLEMENTED | Base config, interceptors |
| Authentication Context | IMPLEMENTED | Minor issue: uses email vs userId |
| Auth Page (/auth) | IMPLEMENTED | Login/register forms, validation |
| Home Page (/) | IMPLEMENTED | Dashboard welcome page |
| Account Page (/account) | IMPLEMENTED | User profile display |
| AppLayout Component | IMPLEMENTED | Header, footer, navigation |
| Router Configuration | IMPLEMENTED | Basic routes, protected routes |
| Users API Module | IMPLEMENTED | Login/register functions |
| Projects API Module | IMPLEMENTED | List/create projects |
| Projects Page (/projects) | NEEDS BUILD | Project management interface |
| Hardware Page (/hardware) | NEEDS BUILD | Resource request & checkout interface |
| Hardware API Module | NEEDS BUILD | API endpoints for hardware operations |
| Error Display | NEEDS ENHANCEMENT | Add user-friendly error messages |
| Loading States | NEEDS ENHANCEMENT | Add spinners for async operations |

### 14.3 Database Status

| Collection | Status | Notes |
|-----------|--------|-------|
| users | IMPLEMENTED | Schema defined, in use |
| projects | IMPLEMENTED | Schema defined, in use |
| hardware_sets | NEEDS BUILD | Schema designed, not created |
| resource_requests | NEEDS BUILD | Schema designed, not created |
| allocations | NEEDS BUILD | Schema designed, not created |
| Indexes | NEEDS BUILD | Should add indexes for performance |

### 14.4 API Endpoints Status

| Endpoint | Status | Notes |
|----------|--------|-------|
| POST /api/auth/login | IMPLEMENTED | User authentication |
| POST /api/users | IMPLEMENTED | User registration |
| GET /api/users | IMPLEMENTED | List users |
| GET /api/users/<id> | IMPLEMENTED | Get user details |
| POST /api/projects | IMPLEMENTED | Create project |
| GET /api/projects | IMPLEMENTED | List projects with filter |
| GET /api/projects/<id> | IMPLEMENTED | Get project details |
| GET /api/hardware | NEEDS BUILD | List hardware sets |
| GET /api/hardware/availability | NEEDS BUILD | Get real-time availability |
| POST /api/hardware/request | NEEDS BUILD | Request resources |
| POST /api/hardware/checkout | NEEDS BUILD | Checkout resources |
| POST /api/hardware/checkin | NEEDS BUILD | Return resources |
| GET /api/hardware/allocations | NEEDS BUILD | View allocation history |
| GET /api/health | IMPLEMENTED | Health check |

---

## 15. Technology Stack

### 15.1 Backend Stack

| Component | Version | Purpose |
|-----------|---------|---------|
| Python | 3.12+ | Runtime language |
| Flask | 3.1.2 | Web framework |
| PyMongo | 4.16.0 | MongoDB driver |
| Flask-CORS | 6.0.2 | Cross-origin support |
| python-dotenv | 1.2.1 | Environment config |

### 15.2 Frontend Stack

| Component | Version | Purpose |
|-----------|---------|---------|
| React | 19.2.0 | UI framework |
| TypeScript | 5.9+ | Type safety |
| Vite | 7.2.4 | Build tool |
| React Router | 7.13.0 | Client-side routing |
| Axios | 1.13.4 | HTTP client |
| Material-UI | 7.3.7 | Component library |
| ESLint | 9.39.1 | Code linting |

### 15.3 Infrastructure Stack

| Component | Version | Purpose |
|-----------|---------|---------|
| MongoDB | 7.0+ | NoSQL database |
| Docker | Latest | Containerization |
| Docker Compose | Latest | Multi-container orchestration |

### 15.4 Development Tools

| Tool | Purpose |
|------|---------|
| pyenv | Python version management |
| npm | Node package manager |
| pip | Python package manager |
| Virtual Environment | Python isolation |
| VS Code | IDE |

---

## 10. Development Workflow

### 17.1 Starting the Application

**Full Stack**:
```bash
./scripts/start_app.sh
# Starts backend (port 5001), frontend (port 5173), and MongoDB
```

**Backend Only**:
```bash
./scripts/start_backend.sh
# Starts Flask development server on port 5001
```

**Frontend Only**:
```bash
./scripts/start_frontend.sh
# Starts Vite dev server on port 5173
```

### 17.2 Development Environment

**Backend Setup**:
```bash
cd backend
python -m venv .venv
source .venv/bin/activate  # or .venv\Scripts\activate on Windows
pip install -e .
```

**Frontend Setup**:
```bash
cd frontend
npm install
npm run dev
```

**MongoDB**:
```bash
# Via Docker Compose
docker-compose up -d mongo

# Or local MongoDB instance
mongod
```

### 17.3 API Testing

**Tools**: Postman, Insomnia, curl, or Thunder Client

**Base URL**: `http://localhost:5001/api`

**Example Login Request**:
```bash
curl -X POST http://localhost:5001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userid":"testuser","password":"testpass"}'
```



---

## 16. Future Enhancements

1. **Authentication Upgrade**: Replace F3/E1 cipher with bcrypt/Argon2
2. **JWT Tokens**: Implement JWT-based authentication instead of session cookies
3. **Admin Dashboard**: Add admin interface for resource management
4. **Advanced Scheduling**: Time-based resource reservations
5. **WebSocket Support**: Real-time availability updates
6. **Mobile App**: Native mobile application
7. **API Documentation**: Swagger/OpenAPI documentation
8. **Automated Testing**: Unit tests and integration tests
9. **Database Optimization**: Connection pooling, query optimization
10. **Monitoring**: Application performance monitoring and logging

---

**Document Version**: 1.1.0  
**Last Updated**: February 10, 2026  
**Status**: Complete with all standard SDD sections  
**Next Review**: After phase 1 development completion
