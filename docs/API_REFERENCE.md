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

---

## Authentication Endpoints

### POST /api/auth/login
Login endpoint for user authentication.

**Status:**  NEED TO BUILD

**Request:**
```json
{
  "userid": "string (unique login identifier)",
  "password": "string (plaintext, hashed on backend)"
}
```

**Response (200):**
```json
{
  "token": "JWT or session token",
  "user": {
    "userId": "string",
    "username": "string",
    "email": "string (optional)"
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
- Password hashing required (bcrypt or hashlib)
- Return JWT token for subsequent requests
- TODO: Determine token expiration time
- TODO: Implement refresh token mechanism

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
    "username": "string",
    "email": "string (optional)"
  }
]
```

### POST /api/users
Create a new user account.

**Status:** [IMPLEMENTED] (needs username field update)

**Request:**
```json
{
  "userid": "string (unique, 3-20 characters)",
  "username": "string (display name, may need uniqueness discussion)",
  "password": "string (plaintext, hashed on backend)"
}
```

**Response (201):**
```json
{
  "userId": "string",
  "username": "string",
  "createdAt": "ISO8601 timestamp"
}
```

**Error Response (409):**
```json
{
  "error": "User already exists",
  "code": 409
}
```

**Implementation Notes:**
- `userid` must be unique (database index required)
- Password must be hashed using bcrypt/hashlib before storage
- TODO: Username uniqueness requirement - need to discuss
- Validate required fields before creating user
- Return sanitized user data (exclude password_hash)

### GET /api/users/{userId}
Get specific user details.

**Status:** [NEEDS IMPLEMENTATION]

**Response (200):**
```json
{
  "userId": "string",
  "username": "string",
  "email": "string (optional)"
}
```

---

## Project Endpoints

### GET /api/projects
List projects for authenticated user.

**Status:** IMPLEMENTED

**Query Parameters:**
- `?ownerUserId={userId}` - Filter by project owner

**Response (200):**
```json
[
  {
    "projectId": "string (unique)",
    "name": "string",
    "description": "string",
    "ownerUserId": "string or ObjectId",
    "createdAt": "ISO8601 timestamp"
  }
]
```

**Implementation Notes:**
- TODO: Clarify whether to use `ownerUserId` (single owner) or `assignedUsers[]` (multiple owners)
- Filter results based on user permissions if needed

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

### POST /api/hardware/checkout
Check out hardware units.

**Status:** NEED TO BUILD

**Request:**
```json
{
  "projectId": "string (project checking out hardware)",
  "hardwareSet": "string (e.g., HWSet1)",
  "units": "integer (quantity to checkout)"
}
```

**Response (200):**
```json
{
  "allocationId": "string",
  "projectId": "string",
  "hardwareSet": "string",
  "units": "integer",
  "type": "checkout",
  "checkedOutAt": "ISO8601 timestamp"
}
```

**Error Response (409):**
```json
{
  "error": "Insufficient hardware available",
  "available": "integer",
  "requested": "integer",
  "code": 409
}
```

**Implementation Notes:**
- Use MongoDB `$inc` operator for atomic updates (prevent overallocation)
- Check available capacity before checkout
- Record allocation in `allocations` collection
- TODO: Add per-unit tracking if needed

### POST /api/hardware/checkin
Check in hardware units.

**Status:**  NEED TO BUILD

**Request:**
```json
{
  "projectId": "string",
  "hardwareSet": "string (e.g., HWSet1)",
  "units": "integer (quantity to return)"
}
```

**Response (200):**
```json
{
  "allocationId": "string",
  "projectId": "string",
  "hardwareSet": "string",
  "units": "integer",
  "type": "checkin",
  "checkedInAt": "ISO8601 timestamp"
}
```

**Implementation Notes:**
- Use MongoDB `$inc` operator for atomic updates
- Validate that project has checked out the requested quantity
- Record return in `allocations` collection
- Update `projectAllotments` in hardware_sets collection

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

## Implementation Status Summary

| Endpoint | Method | Status | Priority |
|----------|--------|--------|----------|
| /api/auth/login | POST | [NEED] | HIGH |
| /api/users | GET | [DONE] | - |
| /api/users | POST | [DONE] (update needed) | HIGH |
| /api/users/{userId} | GET | [BUILD] | MEDIUM |
| /api/projects | GET | [DONE] | - |
| /api/projects | POST | [DONE] (validation) | MEDIUM |
| /api/projects/{projectId} | GET | [BUILD] | MEDIUM |
| /api/hardware | GET | [NEED] | HIGH |
| /api/hardware/availability | GET | [NEED] | HIGH |
| /api/hardware/request | POST | [NEED] | HIGH |
| /api/hardware/checkout | POST | [NEED] | HIGH |
| /api/hardware/checkin | POST | [NEED] | HIGH |
| /api/hardware/allocations | GET | [NEED] | HIGH |
| /api/health | GET | [DONE] | - |

---

## Legend

- [IMPLEMENTED] - Endpoint exists and functions
- [NEEDS BUILD] - Needs to be implemented
- [NEED TO BUILD] - High priority, must be created
- [TODO] - Design decision needed before implementation

---

**Last Updated:** February 10, 2026  
**Version:** 1.0  
**Status:** In Development
