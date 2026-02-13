# SRS Requirements Traceability Matrix

## Requirements Mapping

This document maps Team Project requirements to the updated SRS document to ensure all requirements are captured.

### Stakeholder Needs Coverage

| SN | Requirement | SRS Section | Status |
|----|-----------|-------------|--------|
| SN1 | Secure user accounts and authentication | 3.1 (User Management Module) | Covered |
| SN2 | View hardware resource status by project context | 3.3.2, 3.3.8 (Hardware Availability & Dashboard) | Covered |
| SN3 | Request hardware resources | 3.3.3 (Request Hardware Resources) | Covered |
| SN4 | Checkout and manage allocation resource | 3.3.5, 3.3.7 (Checkout & Allocation Tracking) | Covered |
| SN5 | Check-in hardware resources & validation | 3.3.6 (Check-in Hardware Resources) | Covered |
| SN6 | Deliver Scalable PoC within schedule | 4.6 (Deployment Requirements) | Covered |

### System Requirements Coverage

| SR | Requirement | SRS Section | Status |
|----|-----------|-------------|--------|
| SR-01 | System shall user to create secure accounts | 3.1.1 (User Registration) | Covered |
| SR-02 | System shall encrypt user credentials | 4.2.2 (Password Security) | Covered |
| SR-03 | System shall authenticate user | 3.1.2 (User Authentication) | Covered |
| SR-04 | System shall allow project creation and selection | 3.2.1 (Create Project) | Covered |
| SR-04b | System shall allow a user to join a project using projectId | 3.2.3 (Join Existing Project) | Covered |
| SR-05 | System shall associate projects with owner | 3.2.4, 3.2.5 (Filter by Owner & Metadata) | Covered |
| SR-06 | System shall display hardware capacity | 3.3.1, 3.3.8 (Hardware Inventory & Dashboard) | Covered |
| SR-07 | System shall display hardware availability | 3.3.2, 3.3.8 (Hardware Availability & Dashboard) | Covered |
| SR-08 | System shall allow checkout of hardware | 3.3.5 (Checkout Hardware Resources) | Covered |
| SR-09 | System shall allow check-in of hardware | 3.3.6 (Check-in Hardware Resources) | Covered |
| SR-10 | System shall prevent over-allocation | 3.3.3, 3.3.5 (Validation in Requests & Checkout) | Covered |
| SR-11 | System shall expose REST APIs | 5.3.1 (Backend API Section) | Covered |
| SR-12 | System shall support frontend backend separation (stateless) | 4.5.1 (Frontend-Backend Separation), 4.2.1 (Stateless Backend) | Covered |
| SR-13 | System shall support page-based routing | 4.5.2 (Page-Based Routing) | Covered |
| SR-14 | System shall track frontend and/or action | 3.3.7, 3.3.8 (Allocation History & Dashboard) | Covered |

### Non-Functional Requirements Coverage

| Requirement | SRS Section | Status |
|-----------|-------------|--------|
| Stateless backend - no server-side session for MVP | 4.2.1 (Authentication), 4.5.1 (Frontend-Backend Separation) | Covered |
| Atomic DB updates for availability and allocations | 4.3.4 (Database Atomicity) | Covered |
| Clear error codes (400/401/404/409) | 4.3.3 (Error Handling) | Covered |

### Data Persistence Requirements

The system shall persist the following entities:

| Entity | SRS Section | Status |
|--------|-------------|--------|
| Users | 7.2 (Data Persistence: User Accounts) | Covered |
| Projects | 7.2 (Data Persistence: Projects) | Covered |
| Hardware Sets (HWSet1, HWSet2) | 7.2 (Data Persistence: Hardware Sets) | Covered |
| Resource Requests | 7.2 (Data Persistence: Resource Requests) | Covered |
| Allocations (checkout/check-in) | 7.2 (Data Persistence: Resource Allocations) | Covered |

### Key Implementation Notes

1. **Stateless Backend**: Backend MUST NOT maintain server-side sessions
   - State management moved entirely to client-side React Context
   - All authentication state stored in frontend
   - Backend validates each request independently

2. **Atomic Database Operations**: 
   - Uses MongoDB for direct atomic updates
   - Mongoose transactions or equivalent for multi-step operations
   - Ensures availability calculations remain consistent

3. **Error Code Standards**:
   - `200 OK` - Successful request
   - `400 Bad Request` - Invalid input or validation failure
   - `401 Unauthorized` - Authentication required or failed
   - `404 Not Found` - Resource doesn't exist
   - `409 Conflict` - Resource conflict (duplicate, insufficient capacity, etc.)

4. **Hardware Management**:
   - Two hardware sets: HWSet1 and HWSet2
   - Real-time availability calculated as: `available = totalCapacity - allocatedUnits`
   - Support for request → approval → checkout → check-in workflow

5. **Project Context**:
   - Projects serve as organizational units for hardware requests
   - Users can create and join projects
   - Resources are allocated to projects, not individual users

### API Endpoints Summary

All endpoints are documented in **Section 5.3.1** of the SRS:

**User Management:**
- `GET /api/users` - List users
- `POST /api/users` - Create user

**Project Management:**
- `GET /api/projects` - List projects
- `POST /api/projects` - Create project

**Hardware Management:**
- `GET /api/hardware-sets` - Get hardware definitions
- `GET /api/hardware-availability` - Get real-time availability
- `POST /api/resource-requests` - Request resources
- `POST /api/resource-requests/{id}/checkout` - Checkout resources
- `POST /api/allocations/{id}/checkin` - Check-in resources
- `GET /api/allocations` - View allocation history

---

**Status**: All requirements have been incorporated into the SRS document.

References:
- ECE 382V: Cloud Native App Development Course
- Team Project: Stakeholder Needs and Grading (Version 7.20260116)
- Software Requirements Specification Standard: IEEE Std 830-1998
