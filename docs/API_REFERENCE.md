# API Reference
**Team Project**

Complete API endpoint specification for the Team Project Hardware Checkout System.

## Table of Contents
1. [Authentication Endpoints](#authentication-endpoints)
2. [User Endpoints](#user-endpoints)
3. [Project Endpoints](#project-endpoints)
4. [Hardware Endpoints](#hardware-endpoints)
5. [Health Check](#health-check)
6. [Error Codes](#error-codes)
7. [Response Formats](#response-formats)
8. [Design Decisions](#design-decisions)

---

## Authentication Endpoints

### POST /api/auth/login
Login endpoint for user authentication.

**Status:** IMPLEMENTED

**Request:**
```json
{
  "userId": "string (unique login identifier)",
  "password": "string (plaintext, encrypted on backend)"
}
```

**Response (200):**
```json
{
  "ok": true,
  "message": "Login successful",
  "user": {
    "userId": "string"
  }
}
```

**Error Response (401):**
```json
{
  "error": "Invalid credentials",
  "code": 401
}
```

**Implementation Notes:**
- Only password is encrypted using cyclic cipher algorithm
- userId stored and matched in plaintext
- Returns user object with userId for client use
- Frontend stores full User object in auth context with 24hr session persistence

### POST /api/auth/register
Register a new user account.

**Status:** IMPLEMENTED (Casey's PR)

**Request:**
```json
{
  "userId": "string (unique login identifier)",
  "password": "string (plaintext, encrypted on backend)"
}
```

**Response (201):**
```json
{
  "user": {
    "userId": "string"
  }
}
```

**Error Response (409):**
```json
{
  "error": "userId already exists"
}
```

**Implementation Notes:**
- Only `userId` and `password` required (per professor clarification)
- `userId` must be unique (enforced in code)
- Both `userId` and `password` encrypted using cyclic cipher before storage
- Response returns decrypted userId for frontend auth context
- Frontend passes full User object to `login()` function

---

## User Endpoints

### GET /api/users
List all users.

**Status:** IMPLEMENTED

**Query Parameters:**
- None currently

**Response (200):**
```json
[
  {
    "userId": "string",
    "email": "string (optional)"
  }
]
```

### POST /api/users
Create a new user account.

**Status:** IMPLEMENTED

**Request:**
```json
{
  "userId": "string (unique login identifier)",
  "password": "string (plaintext, encrypted on backend)"
}
```

**Response (201):**
```json
{
  "_id": "string (MongoDB ObjectId)",
  "userId": "string"
}
```

**Error Response (409):**
```json
{
  "error": "userId already exists"
}
```

**Implementation Notes:**
- Only `userId` and `password` required (per professor clarification)
- `userId` must be unique (enforced in code)
- Both `userId` and `password` are encrypted using cyclic cipher before storage
- Return sanitized user data (excludes `password`)

### GET /api/users/{userId}
Get specific user details.

**Status:** [NEEDS IMPLEMENTATION]

**Response (200):**
```json
{
  "userId": "string",
  "email": "string (optional)"
}
```

---

## Project Endpoints

### GET /api/projects
List all projects with associated users and hardware allocations.

**Status:** IMPLEMENTED (response format needs update)

**Query Parameters:**
- `?ownerUserId={userId}` - Filter by project owner

**Response (200):**
```json
[
  {
    "projectId": "string (unique)",
    "name": "string",
    "description": "string",
    "ownerUserId": "string",
    "assignedUsers": ["userid1", "userid2"],
    "hardwareAllocations": [
      {
        "hardwareId": "string",
        "checkedOut": "integer"
      }
    ],
    "createdAt": "ISO8601 timestamp"
  }
]
```

**Implementation Notes:**
- Response includes `assignedUsers[]` - list of user identifiers authorized for the project
- Response includes `hardwareAllocations[]` - hardware currently checked out by this project
- For UI simplicity, return denormalized/assembled data in single call
- Filter results based on user permissions if authentication is implemented

### POST /api/projects
Create a new project.

**Status:** [IMPLEMENTED] (needs validation)

**Request:**
```json
{
  "projectId": "string (unique identifier)",
  "name": "string (required)",
  "description": "string"
}
```

**Response (201):**
```json
{
  "projectId": "string",
  "name": "string",
  "description": "string",
  "ownerUserId": "string (set to authenticated user)",
  "createdAt": "ISO8601 timestamp"
}
```

**Error Response (400):**
```json
{
  "error": "Missing required fields: name",
  "code": 400
}
```

**Implementation Notes:**
- Validate all required fields
- Set `ownerUserId` to authenticated user automatically
- Ensure `projectId` is unique
- TODO: Decide on projectId format (auto-generated vs user-provided)

### GET /api/projects/{projectId}
Get specific project details.

**Status:** NEEDS IMPLEMENTATION

**Response (200):**
```json
{
  "projectId": "string",
  "name": "string",
  "description": "string",
  "ownerUserId": "string",
  "createdAt": "ISO8601 timestamp"
}
```

---

## Hardware Endpoints

### GET /api/hardware
List all available hardware sets.

**Status:** NEED TO BUILD

**Response (200):**
```json
[
  {
    "hardwareId": "string (e.g., HWSet1)",
    "hardwareName": "string",
    "capacity": "integer (total units)",
    "available": "integer (units available)",
    "reserved": "integer (units in use)",
    "projectAllotments": [
      {
        "projectId": "string",
        "checkedOut": "integer (units checked out by this project)"
      }
    ]
  }
]
```

**Implementation Notes:**
- Available = capacity - SUM(projectAllotments[*].checkedOut)
- Returns all hardware regardless of user permission
- TODO: Consider access control per project

### GET /api/hardware/availability
Get real-time availability status.

**Status:** NEED TO BUILD

**Response (200):**
```json
[
  {
    "hardwareId": "string",
    "hardwareName": "string",
    "capacity": "integer",
    "available": "integer",
    "percentageAvailable": "number (0-100)"
  }
]
```

### POST /api/hardware/request
Create a hardware request (for approval workflow).

**Status:** NEED TO BUILD

**Request:**
```json
{
  "projectId": "string (project requesting hardware)",
  "hardwareSet": "string (e.g., HWSet1)",
  "units": "integer (quantity requested)",
  "reason": "string (optional - reason for request)"
}
```

**Response (201):**
```json
{
  "requestId": "string",
  "projectId": "string",
  "hardwareSet": "string",
  "units": "integer",
  "status": "pending|approved|denied",
  "createdAt": "ISO8601 timestamp"
}
```

**Implementation Notes:**
- TODO: Currently designed for auto-approval
- TODO: Implement approval workflow if needed
- Validate available capacity before approval

### POST /api/hardware/{hardwareId}/checkout
Check out hardware units from shared inventory.

**Status:** NEEDS IMPLEMENTATION

**Path Parameters:**
- `hardwareId` - Hardware set identifier (e.g., "HWSet1")

**Request:**
```json
{
  "projectId": "string (project checking out hardware)",
  "qty": "integer (quantity to checkout)",
  "userId": "string (user performing checkout)"
}
```

**Response (200):**
```json
{
  "allocationId": "string",
  "projectId": "string",
  "hardwareId": "string",
  "qty": "integer",
  "type": "checkout",
  "checkedOutAt": "ISO8601 timestamp",
  "available": "integer (remaining availability)"
}
```

**Error Response (409):**
```json
{
  "error": "Insufficient hardware available",
  "available": "integer",
  "requested": "integer"
}
```

**Implementation Notes:**
- Hardware is a shared global inventory (not per-project)
- Backend validates: user is authorized for project, qty > 0, availability >= qty
- Use MongoDB `$inc` operator for atomic updates (prevent race conditions)
- Record allocation in `allocations` collection
- Return updated availability in response

### POST /api/hardware/{hardwareId}/checkin
Check in hardware units back to shared inventory.

**Status:** NEEDS IMPLEMENTATION

**Path Parameters:**
- `hardwareId` - Hardware set identifier (e.g., "HWSet1")

**Request:**
```json
{
  "projectId": "string (project returning hardware)",
  "qty": "integer (quantity to return)",
  "userId": "string (user performing checkin)"
}
```

**Response (200):**
```json
{
  "allocationId": "string",
  "projectId": "string",
  "hardwareId": "string",
  "qty": "integer",
  "type": "checkin",
  "checkedInAt": "ISO8601 timestamp",
  "available": "integer (updated availability)"
}
```

**Implementation Notes:**
- Backend validates: user exists, qty > 0
- Use MongoDB `$inc` operator for atomic updates
- Validate that project has checked out the requested quantity
- Record return in `allocations` collection

### GET /api/hardware/allocations
Get allocation history for a project.

**Status:**  NEED TO BUILD

**Query Parameters:**
- `?projectId={projectId}` - Filter by project (required)

**Response (200):**
```json
[
  {
    "allocationId": "string",
    "projectId": "string",
    "userId": "string",
    "hardwareSet": "string",
    "units": "integer",
    "type": "checkout|checkin",
    "timestamp": "ISO8601 timestamp"
  }
]
```

---

## Health Check

### GET /api/health
System health check endpoint.

**Status:**  IMPLEMENTED

**Response (200):**
```json
{
  "status": "ok",
  "timestamp": "ISO8601 timestamp",
  "version": "string (app version)"
}
```

---

## Error Codes

| Code | Meaning | Example |
|------|---------|---------|
| 200 | OK | Request succeeded |
| 201 | Created | Resource created successfully |
| 400 | Bad Request | Missing/invalid fields |
| 401 | Unauthorized | Invalid credentials or missing token |
| 403 | Forbidden | User lacks permission for resource |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource already exists or insufficient capacity |
| 500 | Server Error | Internal server error |

---

## Response Formats

### Success Response
All successful responses follow this format:

```json
{
  "success": true,
  "data": {
    // Endpoint-specific data
  },
  "timestamp": "ISO8601 timestamp"
}
```

### Error Response
All error responses follow this format:

```json
{
  "success": false,
  "error": "Human-readable error message",
  "code": "HTTP status code",
  "timestamp": "ISO8601 timestamp"
}
```

### Pagination (Future)
For list endpoints that return many results:

```json
{
  "success": true,
  "data": [
    // Array of items
  ],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "total": 100,
    "totalPages": 5
  },
  "timestamp": "ISO8601 timestamp"
}
```

**TODO:** Determine if pagination needed for hardware/allocations lists

---

## Design Decisions

Key architectural decisions agreed upon by the team:

### Authentication Fields
- Only `userId` and `password` required (per professor clarification - no username needed)
- **`userId`** is the canonical login identifier (camelCase in backend response)
- Both `userId` and `password` are encrypted using cyclic cipher before storage
- Auth responses return decrypted `user.userId` for frontend display
- Frontend passes full `User` object to auth context (not just userId string)

### Session Persistence
- Sessions stored in localStorage with base64 encoding
- Session TTL: 24 hours
- Session includes full User object and expiration timestamp

### Hardware Inventory Model
- Hardware is a **shared global inventory** (not per-project)
- Projects check out from a single shared pool
- Availability decreases when any project checks out; increases on check-in
- Backend enforces validation and atomic updates to prevent race conditions
- Checkout/check-in logic must be on backend (not frontend arithmetic)

### Projects Data Model
- Projects include `assignedUsers[]` array of authorized user identifiers
- GET /api/projects returns denormalized data (users + hardware allocations included)
- Single API call should provide all data needed for Projects UI

### Response Consistency
- All auth endpoints return `ok: boolean` and `user` object
- Error responses include descriptive `error` message
- Frontend should not assume field existence without checking

### Database Architecture
- Use 3 separate databases (Users, Projects, Hardware) - NOT 3 collections
- Professor advised this design for Part 2 microservices
- Each entity references others via foreign keys (e.g., projects store user IDs, hardware IDs)
- Mongo Compass recommended for database visualization

### Security Notes
- Both `userId` and `password` MUST be encrypted (per professor confirmation of SR3)
- Current code only encrypts password - needs fix
- Store encrypted values in database, return decrypted userId for UI display
- Remove console.log statements that print credentials (Auth.tsx)
- Remove console.log in Account.tsx (Casey added for debugging)
- Credentials should not be visible in browser DevTools console

---

## Implementation Status Summary

| Endpoint | Method | Status | Priority | Owner |
|----------|--------|--------|----------|-------|
| /api/auth/login | POST | IMPLEMENTED | - | - |
| /api/auth/register | POST | IMPLEMENTED | - | Casey |
| /api/users | GET | IMPLEMENTED | - | - |
| /api/users | POST | IMPLEMENTED | - | - |
| /api/users/{userId} | GET | NEEDS BUILD | MEDIUM | - |
| /api/projects | GET | IMPLEMENTED (update response) | HIGH | Yuri |
| /api/projects | POST | IMPLEMENTED | - | - |
| /api/projects/{projectId} | GET | NEEDS BUILD | MEDIUM | - |
| /api/hardware | GET | NEEDS BUILD | HIGH | Yuri |
| /api/hardware/availability | GET | NEEDS BUILD | HIGH | Yuri |
| /api/hardware/{id}/checkout | POST | NEEDS BUILD | HIGH | Yuri |
| /api/hardware/{id}/checkin | POST | NEEDS BUILD | HIGH | Yuri |
| /api/hardware/allocations | GET | NEEDS BUILD | MEDIUM | - |
| /api/health | GET | IMPLEMENTED | - | - |

---

## Legend

- [IMPLEMENTED] - Endpoint exists and functions
- [NEEDS BUILD] - Needs to be implemented
- [NEED TO BUILD] - High priority, must be created
- [TODO] - Design decision needed before implementation

---

**Last Updated:** February 13, 2026  
**Version:** 1.2  
**Status:** In Development
