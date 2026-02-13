# Software Requirements Specification (SRS)
## Cloud-Native Team Project

**Project Name:** Cloud-Native Team Project  
**Version:** 0.2.0  
**Date:** February 2026  
**Status:** In Development  
**Course:** ECE 382V: Cloud Native App Development  

---

## Table of Contents

1. [Introduction](#1-introduction)
   - [1.1 Purpose](#11-purpose)
   - [1.2 Scope](#12-scope)
   - [1.3 Document Conventions](#13-document-conventions)
   - [1.4 References](#14-references)

2. [Overall Description](#2-overall-description)
   - [2.1 Product Perspective](#21-product-perspective)
   - [2.2 Product Functions](#22-product-functions)
   - [2.3 Stakeholder Needs to System Requirements Mapping](#23-stakeholder-needs-to-system-requirements-mapping)
   - [2.4 System Requirements Summary](#24-system-requirements-summary)
   - [2.5 User Classes and Characteristics](#25-user-classes-and-characteristics)
   - [2.6 Operating Environment](#26-operating-environment)
   - [2.7 Design and Implementation Constraints](#27-design-and-implementation-constraints)

3. [Functional Requirements](#3-functional-requirements)
   - [3.0 System Requirements Detailed](#30-system-requirements-detailed)
   - [3.1 User Management Module](#31-user-management-module)
     - [3.1.1 User Registration](#311-user-registration-fr-um-001)
     - [3.1.2 User Authentication](#312-user-authentication-fr-um-002)
     - [3.1.3 List Users](#313-list-users-fr-um-003)
     - [3.1.4 User Account Management](#314-user-account-management-fr-um-004)
   - [3.2 Project Management Module](#32-project-management-module)
     - [3.2.1 Create Project](#321-create-project-fr-pm-001)
     - [3.2.2 List Projects](#322-list-projects-fr-pm-002)
     - [3.2.3 Join Existing Project](#323-join-existing-project-fr-pm-003)
     - [3.2.4 Filter Projects by Owner](#324-filter-projects-by-owner-fr-pm-004)
     - [3.2.5 Project Metadata](#325-project-metadata-fr-pm-005)
   - [3.3 Hardware Resource Management Module](#33-hardware-resource-management-module)
     - [3.3.1 Hardware Resource Inventory](#331-hardware-resource-inventory-fr-hrm-001)
     - [3.3.2 View Hardware Resource Availability](#332-view-hardware-resource-availability-fr-hrm-002)
     - [3.3.3 Request Hardware Resources](#333-request-hardware-resources-fr-hrm-003)
     - [3.3.4 Approve Resource Requests](#334-approve-resource-requests-fr-hrm-004)
     - [3.3.5 Checkout Hardware Resources](#335-checkout-hardware-resources-fr-hrm-005)
     - [3.3.6 Check-in Hardware Resources](#336-check-in-hardware-resources-fr-hrm-006)
     - [3.3.7 View Resource Allocation History](#337-view-resource-allocation-history-fr-hrm-007)
     - [3.3.8 Hardware Resource Status Dashboard](#338-hardware-resource-status-dashboard-fr-hrm-008)
   - [3.4 Dataset Management Module (Future)](#34-dataset-management-module-future-implementation)
     - [3.4.1 Publish Available Datasets](#341-publish-available-datasets-fr-dsm-001)
     - [3.4.2 Request Available Datasets](#342-request-available-datasets-fr-dsm-002)

4. [Non-Functional Requirements](#4-non-functional-requirements)
   - [4.1 Performance Requirements](#41-performance-requirements)
     - [4.1.1 Response Time](#411-response-time-nfr-p-001)
     - [4.1.2 Throughput](#412-throughput-nfr-p-002)
     - [4.1.3 Scalability](#413-scalability-nfr-p-003)
   - [4.2 Security Requirements](#42-security-requirements)
     - [4.2.1 Authentication](#421-authentication-nfr-s-001)
     - [4.2.2 Password Security](#422-password-security-nfr-s-002)
     - [4.2.3 Data Validation](#423-data-validation-nfr-s-003)
     - [4.2.4 CORS Policy](#424-cors-policy-nfr-s-004)
   - [4.3 Reliability Requirements](#43-reliability-requirements)
     - [4.3.1 Availability](#431-availability-nfr-r-001)
     - [4.3.2 Data Persistence](#432-data-persistence-nfr-r-002)
     - [4.3.3 Error Handling](#433-error-handling-nfr-r-003)
     - [4.3.4 Database Atomicity](#434-database-atomicity-nfr-r-004)
   - [4.4 Maintainability Requirements](#44-maintainability-requirements)
     - [4.4.1 Code Quality](#441-code-quality-nfr-m-001)
     - [4.4.2 Documentation](#442-documentation-nfr-m-002)
     - [4.4.3 Testing](#443-testing-nfr-m-003)
   - [4.5 Usability Requirements](#45-usability-requirements)
     - [4.5.1 Frontend-Backend Separation](#451-frontend-backend-separation-nfr-u-001)
     - [4.5.2 Page-Based Routing](#452-page-based-routing-nfr-u-002)
     - [4.5.3 User Interface](#453-user-interface-nfr-u-003)
   - [4.6 Deployment Requirements](#46-deployment-requirements)
     - [4.6.1 Containerization](#461-containerization-nfr-d-001)
     - [4.6.2 Environment Configuration](#462-environment-configuration-nfr-d-002)
     - [4.6.3 Startup Scripts](#463-startup-scripts-nfr-d-003)

5. [External Interface Requirements](#5-external-interface-requirements)
   - [5.1 User Interfaces](#51-user-interfaces)
     - [5.1.1 Authentication Page](#511-authentication-page-fr-ui-001)
     - [5.1.2 Home Page](#512-home-page-fr-ui-002)
     - [5.1.3 Account Management Page](#513-account-management-page-fr-ui-003)
     - [5.1.4 Protected Routes](#514-protected-routes-fr-ui-004)
   - [5.2 Hardware Interfaces](#52-hardware-interfaces)
   - [5.3 Software Interfaces](#53-software-interfaces)
     - [5.3.1 Backend API](#531-backend-api)
     - [5.3.2 Database Interface](#532-database-interface)
     - [5.3.3 Frontend-Backend Communication](#533-frontend-backend-communication)
   - [5.4 Communication Protocols](#54-communication-protocols)

6. [System Features](#6-system-features)
   - [6.1 Feature: User Registration and Management](#61-feature-user-registration-and-management)
   - [6.2 Feature: Project Management](#62-feature-project-management)
   - [6.3 Feature: Hardware Resource Management](#63-feature-hardware-resource-management)
   - [6.3b Feature: Dataset Management (Future)](#63b-feature-dataset-management-future-enhancement-for-sn3)
   - [6.4 Feature: Protected User Interface](#64-feature-protected-user-interface)
   - [6.5 Feature: RESTful API Services](#65-feature-restful-api-services)
   - [6.6 Feature: Hardware Resource Monitoring Dashboard](#66-feature-hardware-resource-monitoring-dashboard)
   - [6.7 Feature: Containerized Deployment](#67-feature-containerized-deployment)

7. [Data Requirements](#7-data-requirements)
   - [7.1 Logical Data Model](#71-logical-data-model)
   - [7.2 Data Persistence Requirements](#72-data-persistence-requirements)
   - [7.3 Data Validation Rules](#73-data-validation-rules)
   - [7.4 Data Access Requirements](#74-data-access-requirements)
   - [7.5 Data Security](#75-data-security)

8. [Appendices](#8-appendices)
   - [A. Glossary](#a-glossary)
   - [B. Use Cases](#b-use-cases)
   - [C. Assumptions and Constraints](#c-assumptions-and-constraints)
   - [D. Future Enhancements](#d-future-enhancements)

9. [Approval](#9-approval)

10. [References](#references)

---

**Project Name:** Cloud-Native Team Project  
**Version:** 0.2.0  
**Date:** February 2026  
**Status:** In Development  
**Course:** ECE 382V: Cloud Native App Development  

---

## 1. Introduction

### 1.1 Purpose
This Software Requirements Specification (SRS) document describes the functional and non-functional requirements for the Hardware-as-a-Service (HaaS) Proof of Concept (PoC) application. This document serves as a reference for development, testing, and project management activities aligned with ECE 382V: Cloud Native App Development course requirements.

### 1.2 Scope
The HaaS PoC application is a full-stack web platform enabling users to manage accounts, projects, and hardware resources. The system allows users to request, checkout, and check-in hardware resources from available resource sets. It includes a RESTful backend API, a modern React frontend interface, MongoDB database, and containerized deployment infrastructure inspired by the University of Utah POWDER program.

### 1.3 Document Conventions
- **MUST**: Indicates a mandatory requirement
- **SHOULD**: Indicates a preferred requirement
- **MAY**: Indicates an optional requirement
- `Code blocks` represent technical implementation details
- API endpoints are marked with HTTP method (GET, POST, PUT, DELETE)

### 1.4 References
- Backend Flask Boilerplate Documentation
- React + Vite + TypeScript Documentation
- MongoDB Official Documentation
- Docker and Docker Compose Documentation

---

## 2. Overall Description

### 2.1 Product Perspective
The HaaS PoC is a cloud-native, containerized web application designed for managing hardware resources in a distributed system. Users can manage their accounts, organize projects, and request/checkout hardware resources for their projects. The system provides resource tracking, availability monitoring, and checkout/check-in workflows. It operates as a monolithic deployment with separated frontend and backend components, both containerized using Docker.

### 2.2 Product Functions
The system MUST provide the following high-level functions:

1. **User Management** (SN1)
   - Create and maintain secure user accounts
   - Authenticate users with encrypted credentials
   - Support multiple user projects

2. **Project Management** (SN1)
   - Create and manage projects within user accounts
   - Organize resources under project context
   - Track project ownership and metadata

3. **Hardware Resource Management** (SN2, SN3, SN4, SN5)
   - Display hardware resource inventory and capacity (HWSet1, HWSet2)
   - Show resource availability and current utilization
   - Allow users to request available resources
   - Support resource approval workflow
   - Manage resource checkout and check-in operations

4. **Dataset Support** (SN3)
   - Publish and manage available datasets (future enhancement, MVP focuses on hardware)
   - Allow users to request and access published datasets
   - Track dataset availability alongside hardware resources
   - Display real-time hardware status
   - Track resource usage and allocation
   - Report checkout/check-in history

5. **API Services** (All SN)
   - RESTful API for backend operations
   - Health check endpoint for system monitoring
   - CORS-enabled for frontend integration

6. **User Interface** (SR2)
   - Web-based user interface
   - Authentication and authorization UI
   - Account and project management pages
   - Hardware resource browsing and request interface
   - Resource checkout/check-in management
   - Protected routes requiring authentication

### 2.3 Stakeholder Needs to System Requirements Mapping

| Stakeholder Need | Description | Related SRs |
|------------------|-------------|------------|
| SN1 | Create and maintain secure user accounts and projects on the system | SR1, SR2, SR3, SR4, SR5 |
| SN2 | View the status of all hardware resources in the system | SR2, SR5 |
| SN3 | Request available hardware resources and datasets from published sources | SR2, SR5 | Hardware resources (MVP); Datasets (future) |
| SN4 | Once approved, checkout and manage these resources | SR2, SR5 |
| SN5 | Check-in the resources and get status of all hardware resources in the system | SR2, SR5 |
| SN6 | Deliver PoC within schedule constraints, with support for scalability | SR1, SR2 |

### 2.4 System Requirements Summary

| Req ID | Description | SN | Modules |
|--------|-------------|----|---------| 
| SR1 | PoC delivered within budget/schedule with periodic stakeholder updates | SN6 | SDLC, Agile |
| SR2 | Front-end web application for user inputs and outputs | All | Web UI, React.js |
| SR3 | Encryption mechanism for user-id and password | SN1 | Python OOP, Cryptography |
| SR4 | Create new projects or access existing projects | SN1 | Python OOP, Python Modules |
| SR5 | Database for users, projects, and hardware resources | SN2-SN5 | Python, MongoDB |

### 2.5 User Classes and Characteristics

#### 2.3.1 End Users
- **Description**: Individuals who interact with the web application
- **Characteristics**: May have varying technical expertise
- **Expected Frequency**: Daily active users
- **Needs**: Intuitive interface, fast performance, secure authentication

#### 2.3.2 Developers
- **Description**: Development team maintaining the codebase
- **Characteristics**: Technical expertise in full-stack development
- **Expected Frequency**: Continuous development and maintenance
- **Needs**: Clear documentation, modular code, easy debugging

#### 2.3.3 System Administrators
- **Description**: Personnel managing deployment and infrastructure
- **Characteristics**: DevOps and system administration expertise
- **Expected Frequency**: Deployment and operational tasks
- **Needs**: Container orchestration, monitoring, scalability

### 2.4 Operating Environment

#### 2.4.1 Development Environment
- **OS**: macOS, Linux, Windows (with appropriate package managers)
- **Python**: 3.12+ (managed via pyenv)
- **Node.js**: Latest LTS version
- **Database**: MongoDB (local or remote instance)
- **Development Tools**: Virtual environments, npm/pip package managers

#### 2.4.2 Production Environment
- **Containerization**: Docker and Docker Compose
- **Backend Container**: Flask application running on port 5001
- **Frontend**: React application served at port 5173
- **Database**: MongoDB container or remote service
- **Network**: Internal Docker network with volume persistence

#### 2.4.3 Browser Requirements
- **Frontend Compatibility**: Modern browsers supporting ES2020+
- **Recommended**: Chrome, Firefox, Safari, Edge (latest versions)
- **JavaScript**: MUST be enabled
- **Cookies**: MUST be enabled for session management

### 2.5 Design and Implementation Constraints

#### 2.5.1 Technology Stack
- **Backend Framework**: Flask 3.1.2
- **Frontend Framework**: React 19.2.0 with Vite
- **Programming Languages**: Python 3.12+, TypeScript 5.9+, HTML5, CSS3
- **Database**: MongoDB 7.0+
- **Package Managers**: pip (Python), npm (Node.js)
- **Testing**: ESLint for frontend code quality
- **Build Tools**: Vite, TypeScript compiler

#### 2.5.2 Database Schema Constraints
- **Collections Used**:
  - `users`: User account information and credentials
  - `projects`: Project metadata and ownership data
  - Extensible for additional collections as needed

#### 2.5.3 Performance Constraints
- Query results MUST be limited to 200 documents for list endpoints
- CORS MUST be configured for frontend origin
- Flask debug mode enabled during development

---

## 3. System Requirements Detailed

### 3.0 System Requirements Mapping

| ID | Requirement | Stakeholder Need | Description |
|----|-------------|------------------|-------------|
| SR-01 | User Account Creation | SN1 | System shall allow users to create secure accounts |
| SR-02 | User Credentials Encryption | SN1 | System shall encrypt user credentials (userid and password) |
| SR-03 | User Authentication | SN1 | System shall authenticate users with encrypted credentials |
| SR-04 | Project Creation | SN1 | System shall allow project creation and selection |
| SR-04b | Project Joining | SN1 | System shall allow a user to join a project using projectId |
| SR-05 | Project Association | SN1 | System shall associate projects with owner |
| SR-06 | Hardware Capacity Display | SN2 | System shall display hardware capacity (HWSet1, HWSet2) |
| SR-07 | Hardware Availability Display | SN2 | System shall display hardware availability and utilization |
| SR-08 | Hardware Checkout | SN4 | System shall allow checkout and management of allocated resources |
| SR-09 | Hardware Check-in | SN5 | System shall allow check-in of hardware resources |
| SR-10 | Over-allocation Prevention | SN4 | System shall prevent over-allocation of hardware resources |
| SR-11 | REST API Exposure | All | System shall expose REST APIs for frontend-backend communication |
| SR-12 | Frontend-Backend Separation | All | System shall support frontend-backend separation (stateless backend) |
| SR-13 | Page-Based Routing | SN1, SN2, SN4, SN5 | System shall support page-based routing in frontend |
| SR-14 | Action Tracking | SN4, SN5 | System shall track resource allocation and check-in actions |

## 3. Functional Requirements

### 3.1 User Management Module

#### 3.1.1 User Registration (FR-UM-001)
**Requirement**: The system MUST allow users to create a new account.

- **Description**: Users SHALL provide username, userid, and password to register
- **Input**: JSON object with required fields: `username`, `userid`, `password`
- **Validation**:
  - All required fields MUST be present and non-empty
  - `username` MUST be unique across the system
  - `password` MUST be hashed using encryption algorithm (F3/E1 cipher with 3 iterations)
  - `userid` MUST be hashed using same encryption algorithm
- **Output**: HTTP 200 with created user document (MongoDB _id)
- **Error Handling**:
  - Return 400 if required fields are missing
  - Return 400 if request body is not valid JSON
  - Return 409 if username already exists
- **API Endpoint**: `POST /api/users`

#### 3.1.2 User Authentication (FR-UM-002)
**Requirement**: The system MUST authenticate users with username and password.

- **Description**: Frontend auth module SHOULD handle login flow
- **Security**: Credentials MUST be transmitted securely (HTTPS in production)
- **Token Management**: Authentication tokens SHOULD be managed via context
- **Session**: User session MUST persist across page navigations

#### 3.1.3 List Users (FR-UM-003)
**Requirement**: The system MUST provide endpoint to retrieve all users.

- **Description**: Retrieve paginated list of all users in the system
- **Query Parameters**: None (pagination SHOULD be future consideration)
- **Limitations**: Results MUST be limited to 200 users
- **Output**: JSON array of user documents with serialized MongoDB _id
- **API Endpoint**: `GET /api/users`

#### 3.1.4 User Account Management (FR-UM-004)
**Requirement**: Authenticated users MUST be able to view and manage their account.

- **Description**: Account page displaying current user information
- **Features**:
  - Display user profile information
  - MAY allow password changes
  - MAY allow profile updates
- **UI Component**: `/pages/Account.tsx`

### 3.2 Project Management Module

#### 3.2.1 Create Project (FR-PM-001)
**Requirement**: The system MUST allow users to create new projects.

- **Description**: Authenticated users SHALL create projects with metadata
- **Input**: JSON object with required fields: `projectId`, `name`, `description`, `ownerUserId`
- **Validation**:
  - `projectId` MUST be unique and non-empty
  - `name` MUST be non-empty and descriptive
  - `description` MUST be non-empty
  - `ownerUserId` SHOULD be current authenticated user's ID
  - All fields MUST be validated before database insertion
- **Output**: HTTP 200 with created project document
- **Error Handling**:
  - Return 400 if required fields are missing or empty
  - Return 400 if request body is not valid JSON
  - Return 409 if projectId already exists
- **API Endpoint**: `POST /api/projects`

#### 3.2.2 List Projects (FR-PM-002)
**Requirement**: The system MUST provide endpoint to retrieve projects.

- **Description**: Retrieve user's projects with optional filtering by owner
- **Query Parameters**: 
  - `ownerUserId` (optional): Filter projects by owner
- **Limitations**: Results MUST be limited to 200 projects
- **Output**: JSON array of project documents
- **API Endpoint**: `GET /api/projects`

#### 3.2.3 Join Existing Project (FR-PM-003)
**Requirement**: Users MUST be able to join existing projects using projectId (SR-04b).

- **Description**: User can access existing project by providing valid projectId
- **Input**: projectId (provided by project owner)
- **Validation**:
  - `projectId` MUST exist in system
  - User credentials MUST be valid
  - User SHOULD NOT already be member (or allow multiple access)
- **Output**: User gains access to project context
- **Error Handling**:
  - Return 404 if projectId not found
  - Return 400 if projectId is empty/invalid
- **UI Component**: Project selection/joining interface
- **Note**: SHOULD track project members/collaborators in future implementation

#### 3.2.4 Filter Projects by Owner (FR-PM-004)
**Requirement**: The system MUST support filtering projects by owner userId (SR-05).

- **Description**: When `ownerUserId` query parameter provided, MUST return only projects owned by that user
- **Implementation**: MongoDB query with `{"ownerUserId": ownerUserId}` filter
- **Use Cases**: User dashboard showing personal projects

#### 3.2.5 Project Metadata (FR-PM-005)
**Requirement**: Each project MUST store and expose relevant metadata.

- **Required Fields**:
  - `projectId`: Unique identifier (string)
  - `name`: Human-readable project name
  - `description`: Project description
  - `ownerUserId`: ID of project owner
- **Auto-Generated**: `_id` (MongoDB ObjectId), `createdAt` (timestamp)

### 3.3 Hardware Resource Management Module

#### 3.3.1 Hardware Resource Inventory (FR-HRM-001)
**Requirement**: The system MUST maintain and display hardware resource inventory across multiple hardware sets.

- **Description**: Define available hardware resource sets (HWSet1, HWSet2) with specifications
- **Hardware Sets**: System MUST support at least two distinct hardware sets
- **Fields per Hardware Set**:
  - `hwsetId`: Unique identifier (HWSet1, HWSet2)
  - `name`: Human-readable hardware set name
  - `totalCapacity`: Total units available in this hardware set
  - `description`: Hardware set description and specifications
  - `specifications`: Technical specifications (CPU, RAM, storage, etc.)
- **Storage**: Hardware set definitions MUST be stored in MongoDB
- **API Endpoint**: `GET /api/hardware-sets`
- **Output**: JSON array of hardware set objects with capacity information

#### 3.3.2 View Hardware Resource Availability (FR-HRM-002)
**Requirement**: The system MUST display real-time availability of each hardware resource set.

- **Description**: Show available vs. allocated units for each hardware set
- **Calculation**: `available = totalCapacity - allocated`
- **Fields to Display**:
  - `hwsetId`: Hardware set identifier
  - `totalCapacity`: Total units in set
  - `allocatedUnits`: Currently checked-out units
  - `availableUnits`: Units available for checkout
  - `utilizationPercentage`: (allocatedUnits / totalCapacity) × 100
- **Real-time Updates**: Availability MUST update when resources are checked out/in
- **API Endpoint**: `GET /api/hardware-availability`
- **Output**: JSON array with availability status for all hardware sets

#### 3.3.3 Request Hardware Resources (FR-HRM-003)
**Requirement**: Users MUST be able to request hardware resources for their projects.

- **Description**: Submit resource request with quantity and project context
- **Input Fields**:
  - `projectId`: Associated project identifier
  - `hwsetId`: Which hardware set (HWSet1, HWSet2)
  - `quantityRequested`: Number of units requested
  - `requestDetails`: Additional request information (optional)
  - `requesterUserId`: ID of requesting user
- **Validation**:
  - `quantityRequested` MUST be positive integer ≤ available units
  - `projectId` MUST exist and user MUST own project
  - `hwsetId` MUST be valid (HWSet1 or HWSet2)
- **Status**: Request created with status `PENDING` (awaiting approval)
- **Output**: Request document with unique request ID
- **Error Handling**:
  - Return 400 if validation fails
  - Return 409 if insufficient resources available
  - Return 404 if project not found
- **API Endpoint**: `POST /api/resource-requests`

#### 3.3.4 Approve Resource Requests (FR-HRM-004)
**Requirement**: System MUST support resource request approval workflow (FUTURE: may be admin-only).

- **Description**: Approve pending resource requests and transition to checkout-ready state
- **Current Implementation**: Auto-approval upon request (future: admin approval workflow)
- **Status Transition**: `PENDING` → `APPROVED` → `CHECKED_OUT`
- **Validation**: Cannot approve request if insufficient resources available
- **API Endpoint**: `POST /api/resource-requests/{requestId}/approve` (future implementation)

#### 3.3.5 Checkout Hardware Resources (FR-HRM-005)
**Requirement**: Users MUST be able to checkout approved hardware resources.

- **Description**: Finalize resource allocation and allocate units to user/project
- **Input**:
  - `requestId`: Approved resource request ID
  - `projectId`: Target project for resource allocation
- **Process**:
  - Review request is APPROVED status
  - Verify resources still available
  - Allocate units to project
  - Record checkout timestamp
  - Transition status to `CHECKED_OUT`
  - Update hardware set allocated units
- **Output**: Checkout confirmation with allocation details
- **Allocation Document**:
  - `allocationId`: Unique allocation identifier
  - `projectId`: Associated project
  - `hwsetId`: Which hardware set
  - `unitsAllocated`: Number of units
  - `checkoutTime`: ISO 8601 timestamp
  - `status`: CHECKED_OUT
  - `ownerUserId`: User who checked out resources
- **Error Handling**:
  - Return 400 if request not approved
  - Return 409 if resources no longer available
- **API Endpoint**: `POST /api/resource-requests/{requestId}/checkout`

#### 3.3.6 Check-in Hardware Resources (FR-HRM-006)
**Requirement**: Users MUST be able to check-in hardware resources and free up capacity.

- **Description**: Return allocated resources and update availability
- **Input**:
  - `allocationId`: ID of allocation to return
  - `projectId`: Project releasing resources
- **Process**:
  - Verify allocation exists and belongs to user
  - Record check-in timestamp
  - Return units to hardware set
  - Update allocated units counter
  - Transition status to `CHECKED_IN`
- **Output**: Check-in confirmation with return details
- **Timestamp Tracking**: Record check-in time for usage tracking
- **Error Handling**:
  - Return 404 if allocation not found
  - Return 403 if user unauthorized
- **API Endpoint**: `POST /api/allocations/{allocationId}/checkin`

#### 3.3.7 View Resource Allocation History (FR-HRM-007)
**Requirement**: Users MUST be able to view their resource checkout/check-in history.

- **Description**: Display all resource allocations and transactions for user's projects
- **Query Parameters**:
  - `projectId` (optional): Filter by specific project
  - `userId`: Current authenticated user ID
- **Output**: List of allocation documents with timestamps
- **Fields Displayed**:
  - Hardware set name and type
  - Quantity allocated
  - Checkout and check-in times
  - Current allocation status
  - Total duration of allocation
- **API Endpoint**: `GET /api/allocations?projectId=<id>`

#### 3.3.8 Hardware Resource Status Dashboard (FR-HRM-008)
**Requirement**: System MUST provide comprehensive hardware status overview.

- **Description**: Dashboard displaying all hardware sets with capacity, availability, and utilization
- **Display Elements**:
  - Hardware set inventory table or cards
  - Real-time availability metrics
  - Current usage by projects
  - Allocation activity timeline (recent checkouts/check-ins)
- **UI Component**: Resource Management section of frontend
- **Refresh Rate**: Display SHOULD refresh in real-time or on user action
- **Responsive Design**: MUST be accessible on desktop and mobile

### 3.4 Dataset Management Module (Future Implementation)

#### 3.4.1 Publish Available Datasets (FR-DSM-001)
**Requirement**: The system MAY support publishing available datasets (SN3 future enhancement).

- **Description**: Make datasets available for user request and download
- **Dataset Metadata**: Include dataset name, description, size, format, availability
- **Access Control**: Datasets may be public or private (owner-restricted)
- **Current MVP Status**: DEFERRED - focus on hardware resources first
- **Future Implementation**: Add dataset management interface and APIs
- **Note**: This feature is identified for future implementation beyond MVP

#### 3.4.2 Request Available Datasets (FR-DSM-002)
**Requirement**: Users MAY be able to request datasets (SN3 future enhancement).

- **Description**: Allow users to request or download published datasets
- **Correlation**: Dataset requests may be linked to projects for tracking
- **Current MVP Status**: DEFERRED - hardware resource requests are priority
- **Future Implementation**: Add dataset request workflow alongside hardware requests

---

## 4. Non-Functional Requirements

### 4.1 Performance Requirements (NFR-P)

#### 4.1.1 Response Time (NFR-P-001)
- **Requirement**: API responses MUST complete within 2 seconds for normal operations
- **Database**: MongoDB queries MUST complete within 500ms for list operations
- **Frontend**: Page load MUST complete within 3 seconds on standard network

#### 4.1.2 Throughput (NFR-P-002)
- **Requirement**: System MUST support minimum 100 concurrent users
- **Database Connections**: MongoDB connection pool SHOULD be configured appropriately
- **Query Optimization**: List endpoints limited to 200 results

#### 4.1.3 Scalability (NFR-P-003)
- **Horizontal Scaling**: Docker containers SHOULD be easily replicable
- **Database Scaling**: MongoDB SHOULD be deployable as standalone or ReplicaSet
- **Load Balancing**: Future deployments MAY include load balancing

### 4.2 Security Requirements (NFR-S)

#### 4.2.1 Authentication (NFR-S-001)
- **Requirement**: All user-sensitive operations MUST require authentication
- **Implementation**: React Context API for session management
- **Token Storage**: Authentication state MUST be stored securely in frontend
- **Stateless Backend**: Backend MUST NOT maintain server-side sessions (SR-12)
- **Session Management**: Authentication state MUST be managed client-side with secure mechanisms

#### 4.2.2 Password Security (NFR-S-002)
- **Requirement**: Passwords MUST be encrypted before storage (SR-02)
- **Algorithm**: F3/E1 cipher with 3 iterations (current implementation)
- **Future**: SHOULD migrate to industry-standard algorithms (bcrypt, Argon2)
- **Transmission**: HTTPS MUST be used in production
- **Encryption Scope**: Both `userid` and `password` fields MUST be encrypted (SR-03)

#### 4.2.3 Data Validation (NFR-S-003)
- **Input Validation**: All API inputs MUST be validated for type and content
- **SQL/NoSQL Injection**: Application SHOULD use parameterized queries (PyMongo handles this)
- **CORS**: MUST be configured to allow only trusted origins

#### 4.2.4 CORS Policy (NFR-S-004)
- **Requirement**: Backend MUST enforce Cross-Origin Resource Sharing
- **Default**: `CORS_ORIGINS=http://localhost:5173` (frontend default)
- **Configuration**: MUST be environment-specific (dev, staging, production)

### 4.3 Reliability Requirements (NFR-R)

#### 4.3.1 Availability (NFR-R-001)
- **Uptime Target**: 99% availability during development
- **Graceful Degradation**: System SHOULD degrade gracefully on database unavailability
- **Health Checks**: `/api/health` endpoint MUST return system status

#### 4.3.2 Data Persistence (NFR-R-002)
- **Database**: MongoDB data MUST persist across container restarts
- **Volumes**: Docker volumes MUST be configured for `mongo_data`
- **Backup**: SHOULD implement regular backup procedures

#### 4.3.3 Error Handling (NFR-R-003)
- **Graceful Failures**: API MUST return appropriate HTTP status codes
- **Standard Error Codes**:
  - `200 OK`: Successful request
  - `400 Bad Request`: Invalid input or malformed request
  - `401 Unauthorized`: Authentication required or failed
  - `404 Not Found`: Resource does not exist
  - `409 Conflict`: Resource conflict (e.g., duplicate projectId, insufficient resources)
- **Error Response Format**: MUST return JSON with error message and status code
- **Logging**: Backend SHOULD log errors for debugging (Flask DEBUG mode)
- **User Feedback**: Frontend MUST display user-friendly error messages

#### 4.3.4 Database Atomicity (NFR-R-004)
- **Atomic Updates**: Database updates for availability and allocations MUST be atomic
- **Consistency**: Resource availability calculations MUST remain consistent across concurrent operations
- **Transaction Support**: SHOULD use MongoDB transactions where applicable
- **Allocation Integrity**: Checkout and check-in operations MUST not result in lost or duplicated allocations

### 4.4 Maintainability Requirements (NFR-M)

#### 4.4.1 Code Quality (NFR-M-001)
- **Requirement**: Code MUST follow style guidelines and best practices
- **Backend**: Python PEP 8 style guidelines
- **Frontend**: ESLint configuration MUST be enforced
- **Type Safety**: TypeScript MUST be used for type checking

#### 4.4.2 Documentation (NFR-M-002)
- **Code Documentation**: Functions SHOULD have docstrings/comments
- **README Files**: Each module SHOULD have comprehensive README
- **API Documentation**: Endpoint behavior MUST be documented

#### 4.4.3 Testing (NFR-M-003)
- **Unit Tests**: SHOULD implement unit tests for business logic
- **Integration Tests**: SHOULD test API endpoints
- **Linting**: Frontend linting MUST pass before deployment

### 4.5 Usability Requirements (NFR-U)

#### 4.5.1 Frontend-Backend Separation (NFR-U-001)
- **Requirement**: Frontend and backend MUST be fully separated with no shared state (SR-12)
- **Stateless Backend**: Backend MUST NOT maintain server-side sessions
- **Frontend State Management**: All session and authentication state MUST be managed on client-side
- **API Communication**: All frontend-backend communication MUST be via REST API (SR-11)
- **Independent Deployment**: Frontend and backend MUST be independently deployable

#### 4.5.2 Page-Based Routing (NFR-U-002)
- **Requirement**: Frontend MUST support page-based routing (SR-13)
- **Router Implementation**: React Router MUST be used for client-side routing
- **Protected Routes**: Routes requiring authentication MUST be protected via ProtectedRoute component
- **Supported Routes**: 
  - `/auth` - Authentication page
  - `/home` or `/` - Home/dashboard page
  - `/account` - User account management page
  - Additional feature pages (resource management, allocation history, etc.)
- **Browser History**: Browser back/forward buttons MUST work correctly with routing

#### 4.5.3 User Interface (NFR-U-003)
- **Responsive Design**: UI MUST be responsive across devices
- **Intuitive Navigation**: Navigation MUST be clear and intuitive
- **Accessibility**: SHOULD follow WCAG 2.1 guidelines (Level AA)

### 4.6 Deployment Requirements (NFR-D)

#### 4.6.1 Containerization (NFR-D-001)
- **Docker**: Application MUST be containerized with Dockerfile
- **Docker Compose**: Multi-container setup MUST use docker-compose.yml
- **Volumes**: MUST define persistent volumes for MongoDB data

#### 4.6.2 Environment Configuration (NFR-D-002)
- **Environment Variables**: MUST use .env file for configuration
- **Required Variables**:
  - `MONGO_URI`: MongoDB connection string
  - `MONGO_DB`: Database name
  - `CORS_ORIGINS`: Authorized frontend origins
  - `FLASK_DEBUG`: Debug mode flag
- **Defaults**: Sensible defaults SHOULD be provided

#### 4.6.3 Startup Scripts (NFR-D-003)
- **Automated Scripts**: MUST provide startup scripts in `scripts/` directory
  - `start_app.sh`: Start entire application
  - `start_backend.sh`: Start backend service
  - `start_frontend.sh`: Start frontend service
- **Script Functionality**: MUST handle virtual environment activation and dependency checks

---

## 5. External Interface Requirements

### 5.1 User Interfaces

#### 5.1.1 Authentication Page (FR-UI-001)
- **Route**: `/auth`
- **Component**: `/src/pages/Auth.tsx`
- **Features**:
  - Login form with username/password
  - Registration form with user details
  - Form validation
  - Error message display

#### 5.1.2 Home Page (FR-UI-002)
- **Route**: `/` or `/home`
- **Component**: `/src/pages/Home.tsx`
- **Features**:
  - Dashboard overview
  - Welcome message for authenticated users
  - Navigation to other sections

#### 5.1.3 Account Management Page (FR-UI-003)
- **Route**: `/account`
- **Component**: `/src/pages/Account.tsx`
- **Features**:
  - Display user profile information
  - Account settings and options
  - Protected route (authentication required)

#### 5.1.4 Protected Routes (FR-UI-004)
- **Route Protection**: `/account` and other sensitive routes MUST require authentication
- **Implementation**: `ProtectedRoute.tsx` component
- **Behavior**: Redirect unauthenticated users to `/auth`

### 5.2 Hardware Interfaces
- **Server**: Standard x86-64 architecture
- **Memory**: 2GB minimum RAM per container
- **Storage**: Scalable based on MongoDB data size

### 5.3 Software Interfaces

#### 5.3.1 Backend API

##### Health Check Endpoint
- **Endpoint**: `GET /api/health`
- **Response**: JSON status object
- **Purpose**: System status monitoring

##### User Endpoints
- **List Users**: `GET /api/users`
  - Response: Array of user objects (limited to 200)
  
- **Create User**: `POST /api/users`
  - Request Body: `{username, userid, password}`
  - Response: User object with inserted _id
  - Status Codes: 200, 400, 409

#### Project Endpoints
- **List Projects**: `GET /api/projects?ownerUserId=<userId>`
  - Query Parameters: `ownerUserId` (optional)
  - Response: Array of project objects (limited to 200)
  
- **Create Project**: `POST /api/projects`
  - Request Body: `{projectId, name, description, ownerUserId}`
  - Response: Project object with inserted _id
  - Status Codes: 200, 400, 409

##### Hardware Resource Endpoints
- **List Hardware Sets**: `GET /api/hardware-sets`
  - Response: Array of all hardware set definitions
  - Includes: hwsetId, name, totalCapacity, specifications, description

- **Get Hardware Availability**: `GET /api/hardware-availability`
  - Response: Array of hardware sets with real-time availability
  - Includes: hwsetId, totalCapacity, allocatedUnits, availableUnits, utilizationPercentage

- **Request Hardware Resources**: `POST /api/resource-requests`
  - Request Body: `{projectId, hwsetId, quantityRequested, requestDetails}`
  - Response: Resource request object with requestId and status
  - Status Codes: 200, 400, 409 (insufficient resources)

- **Checkout Resources**: `POST /api/resource-requests/{requestId}/checkout`
  - Request Body: `{projectId}`
  - Response: Allocation object with allocationId and checkoutTime
  - Status Codes: 200, 400, 409

- **Check-in Resources**: `POST /api/allocations/{allocationId}/checkin`
  - Response: Confirmation with checkinTime and updated availability
  - Status Codes: 200, 404, 403

- **View Allocations**: `GET /api/allocations?projectId=<projectId>`
  - Query Parameters: `projectId` (optional)
  - Response: Array of allocation documents with status and timestamps
  - Filtered by authenticated user's projects

- **Get Request Status**: `GET /api/resource-requests/{requestId}`
  - Response: Resource request document with current status
  - Status Codes: 200, 404

#### 5.3.2 Database Interface
- **Database Type**: MongoDB 7.0+
- **Connection**: PyMongo client with connection pooling
- **Collections**:
  - `users`: User account documents with encrypted credentials
  - `projects`: User project definitions
  - `hardware_sets`: Hardware resource definitions (HWSet1, HWSet2, etc.)
  - `resource_requests`: User resource requests with status tracking
  - `allocations`: Active and historical resource allocations
- **Indexing**: SHOULD include indexes on frequently queried fields (ownerUserId, hwsetId, status)

#### 5.3.3 Frontend-Backend Communication
- **Protocol**: HTTP/HTTPS (REST)
- **Format**: JSON
- **Client Library**: Axios for HTTP requests
- **CORS**: Flask-CORS with configurable origins
- **Authentication**: Token/session via React Context

### 5.4 Communication Protocols
- **REST API**: HTTP 1.1
- **WebSocket**: Not currently used (future consideration)
- **Database**: MongoDB wire protocol
- **Container Networking**: Docker network bridge

---

## 6. System Features

### 6.1 Feature: User Registration and Management
- **Description**: Complete user lifecycle management with secure authentication
- **Scope**: Users, Account management, Authentication, Security
- **Priority**: Critical
- **Stakeholder Needs**: SN1
- **Functional Requirements**: FR-UM-001, FR-UM-002, FR-UM-003, FR-UM-004

### 6.2 Feature: Project Management
- **Description**: Create, retrieve, and manage projects with ownership tracking and joining
- **Scope**: Projects, Filtering, Metadata, Context, Collaboration
- **Priority**: Critical
- **Stakeholder Needs**: SN1
- **Functional Requirements**: FR-PM-001, FR-PM-002, FR-PM-003, FR-PM-004, FR-PM-005

### 6.3 Feature: Hardware Resource Management
- **Description**: Core feature for managing hardware inventory, availability, and allocations
- **Scope**: Hardware sets, Resource requests, Checkout/Check-in, Availability tracking
- **Priority**: Critical
- **Stakeholder Needs**: SN2, SN3 (hardware aspect), SN4, SN5
- **Functional Requirements**: FR-HRM-001 through FR-HRM-008
- **Key Workflows**:
  - Request resources by quantity and hardware set
  - Approve/allocate resources to projects
  - Checkout resources for active use
  - Check-in resources to free capacity
  - View real-time hardware availability

### 6.3b Feature: Dataset Management (Future Enhancement for SN3)
- **Description**: Support for publishing and requesting datasets from published sources
- **Scope**: Dataset publishing, discovery, request workflows
- **Priority**: Medium (addresses SN3 - deferred for future implementation)
- **Stakeholder Needs**: SN3 (dataset aspect)
- **Functional Requirements**: FR-DSM-001, FR-DSM-002
- **MVP Status**: Out of scope (hardware resources are MVP focus)
- **Future Implementation**: Add dataset management alongside hardware resource management

### 6.4 Feature: Protected User Interface
- **Description**: Route protection and authentication-based access control
- **Scope**: Frontend routing, Access control, Authorization
- **Priority**: High
- **Stakeholder Needs**: SN1
- **Functional Requirements**: FR-UI-001 to FR-UI-004

### 6.5 Feature: RESTful API Services
- **Description**: Comprehensive API for CRUD operations across all modules
- **Scope**: API design, Response formatting, Error handling
- **Priority**: Critical
- **Stakeholder Needs**: All
- **Non-Functional Requirements**: NFR-P, NFR-S

### 6.6 Feature: Hardware Resource Monitoring Dashboard
- **Description**: Real-time visualization of hardware status and utilization
- **Scope**: Dashboard, Status display, Utilization metrics
- **Priority**: High
- **Stakeholder Needs**: SN2, SN5
- **Functional Requirements**: FR-HRM-002, FR-HRM-008

### 6.7 Feature: Containerized Deployment
- **Description**: Docker and Docker Compose setup for environment consistency and scalability
- **Scope**: Container orchestration, Environment management, Scaling support
- **Priority**: High
- **Stakeholder Needs**: SN6
- **Non-Functional Requirements**: NFR-D

---

## 7. Data Requirements

### 7.1 Logical Data Model

#### Users Collection
```
{
  _id: ObjectId,
  username: String (unique),
  userid: String (encrypted),
  password: String (encrypted),
  createdAt: Date (auto)
}
```

#### Projects Collection
```
{
  _id: ObjectId,
  projectId: String (unique),
  name: String,
  description: String,
  ownerUserId: String,
  createdAt: Date (auto)
}
```

#### Hardware Sets Collection
```
{
  _id: ObjectId,
  hwsetId: String (unique, e.g., "HWSet1", "HWSet2"),
  name: String,
  description: String,
  totalCapacity: Number (total available units),
  specifications: {
    cpu: String (e.g., "Intel Xeon"),
    ram: String (e.g., "64GB"),
    storage: String (e.g., "1TB SSD"),
    additionalSpecs: Array
  },
  createdAt: Date (auto)
}
```

#### Resource Requests Collection
```
{
  _id: ObjectId,
  requestId: String (unique),
  projectId: String (reference to Projects),
  hwsetId: String (reference to Hardware Sets),
  requesterUserId: String,
  quantityRequested: Number,
  requestDetails: String (optional),
  status: String (enum: "PENDING", "APPROVED", "CHECKED_OUT", "CHECKED_IN", "REJECTED"),
  requestTimestamp: Date (auto),
  approvalTimestamp: Date (nullable),
  createdAt: Date (auto)
}
```

#### Resource Allocations Collection
```
{
  _id: ObjectId,
  allocationId: String (unique),
  projectId: String (reference to Projects),
  hwsetId: String (reference to Hardware Sets),
  ownerUserId: String,
  unitsAllocated: Number,
  checkoutTime: Date (ISO 8601),
  checkinTime: Date (nullable, ISO 8601),
  status: String (enum: "CHECKED_OUT", "CHECKED_IN"),
  requestId: String (reference to Resource Requests),
  createdAt: Date (auto)
}
```

#### Hardware Availability View (Computed)
```
{
  hwsetId: String,
  name: String,
  totalCapacity: Number,
  allocatedUnits: Number (sum of unitsAllocated from active allocations),
  availableUnits: Number (totalCapacity - allocatedUnits),
  utilizationPercentage: Number,
  lastUpdated: Date
}
```

### 7.2 Data Persistence Requirements

The system MUST persist the following entities in MongoDB:

1. **User Accounts** - User credentials and account information
   - Collection: `users`
   - Fields: username (unique), encrypted userid, encrypted password, createdAt
   - Used by: SR-01, SR-02, SR-03

2. **Projects** - Project definitions and ownership information
   - Collection: `projects`
   - Fields: projectId (unique), name, description, ownerUserId, createdAt
   - Used by: SR-04, SR-04b, SR-05

3. **Hardware Sets** - Hardware resource definitions and specifications
   - Collection: `hardware_sets`
   - Fields: hwsetId (unique), name, description, totalCapacity, specifications
   - Used by: SR-06, SR-07, SR-08, SR-09, SR-10

4. **Resource Requests** - User requests for hardware resources
   - Collection: `resource_requests`
   - Fields: requestId (unique), projectId, hwsetId, requesterUserId, quantityRequested, status, timestamps
   - Used by: SR-03, SR-08, SR-14

5. **Resource Allocations** - Checkout/check-in tracking and allocation history
   - Collection: `allocations`
   - Fields: allocationId (unique), projectId, hwsetId, ownerUserId, unitsAllocated, checkoutTime, checkinTime, status
   - Used by: SR-08, SR-09, SR-10, SR-14

### 7.3 Data Validation Rules
- **Username**: MUST be unique, non-empty, minimum 3 characters (recommendation)
- **ProjectId**: MUST be unique, non-empty
- **Passwords**: MUST be encrypted, minimum 6 characters (recommendation)
- **Timestamps**: MUST be ISO 8601 format

### 7.4 Data Access Requirements
- **Read Access**: All authenticated users can list data (limited to 200 results)
- **Write Access**: Users can only create owned resources
- **Update/Delete**: SHOULD be restricted to resource owners (future implementation)

### 7.5 Data Security
- **Encryption at Rest**: MongoDB encryption SHOULD be enabled in production
- **Encryption in Transit**: HTTPS MUST be used for all data transmission
- **Password Hashing**: Current implementation SHOULD be upgraded to bcrypt/Argon2

---

## 8. Appendices

### A. Glossary
- **CORS**: Cross-Origin Resource Sharing
- **Docker**: Container platform for application deployment
- **MongoDB**: Document-oriented NoSQL database
- **PyMongo**: Python driver for MongoDB
- **React**: JavaScript library for building user interfaces
- **TypeScript**: Typed superset of JavaScript
- **Vite**: Frontend build tool and development server

### B. Use Cases

#### Use Case: User Registration
1. User navigates to authentication page
2. User enters username, userid, and password
3. System validates input and checks username uniqueness
4. System encrypts credentials and stores in database
5. System returns confirmation with user ID

#### Use Case: Creating a Project
1. Authenticated user navigates to project creation form
2. User enters projectId, name, description
3. System validates input and checks projectId uniqueness
4. System stores project with owner reference
5. System displays success message and project details

#### Use Case: Checkout Hardware Resources
1. Authenticated user navigates to resource management interface
2. User views available hardware sets (HWSet1, HWSet2) and real-time availability
3. User creates resource request specifying quantity desired
4. System validates request (sufficient capacity, valid project context)
5. System transitions request to APPROVED state
6. User initiates checkout process
7. System allocates resources and records checkout timestamp
8. System updates availability metrics in real-time
9. User receives allocation confirmation with resource details

#### Use Case: Check-in Hardware Resources
1. User with active allocation navigates to resource management
2. System displays current allocations for user's projects
3. User selects allocation to return
4. User initiates check-in process
5. System validates ownership and status
6. System records check-in timestamp and frees resources
7. System updates availability metrics
8. System updates allocation status to CHECKED_IN
9. Resources become available for other users

#### Use Case: View Hardware Status and Availability
1. User accesses hardware resource dashboard
2. System queries hardware sets collection
3. System computes real-time availability (totalCapacity - allocated)
4. System displays utilization percentages
5. System shows recent checkout/check-in activity
6. System refreshes metrics on user action or periodic interval

### C. Assumptions and Constraints
- **Assumptions**:
  - Single instance deployment for PoC (no distributed deployment)
  - MongoDB available and accessible
  - Modern browser capabilities available (ES2020+)
  - Python 3.12+ runtime available
  - Node.js LTS available
  - Team can meet scheduled project deadlines
  - Hardware sets (HWSet1, HWSet2) are static during PoC phase
  - Sufficient computational resources for containerized deployment
  - Single admin or trust-based approval system (no formal RBAC in MVP)

- **Course Constraints**:
  - Project must be delivered within course schedule
  - Team size: 5-6 students
  - One submission per team per phase
  - Support for scalability (containerized architecture)
  - Use of course modules: Web UI (React), Python OOP, SDLC/Agile

- **Technical Constraints**:
  - 200 document limit per list query (MongoDB pagination consideration)
  - Single encryption algorithm for passwords (F3/E1 cipher - future upgrade recommended)
  - Monolithic architecture (no microservices in MVP)
  - No external hardware integrations (mock hardware status)
  - Resource allocations are logical (no actual hardware control)
  - Auto-approval workflow (no formal approval process)

### D. Future Enhancements
1. **Dataset Management** (SN3 enhancement): Add support for publishing and requesting datasets from published sources (as per POWDER inspiration and SN3)
2. **Admin Approval Workflow**: Implement admin review and approval of resource requests (currently auto-approve)
3. **Advanced Resource Scheduling**: Support time-based reservations and scheduling of hardware resources
5. **Reservation Cancellation**: Allow users to cancel pending or approved resource requests
6. **Resource Reservation Expiry**: Auto-check-in resources if not checked out within SLA timeframe
7. **Usage Reports and Analytics**: Generate detailed usage reports and analytics dashboards
8. **WebSocket Support**: Real-time updates for resource status via WebSocket connections
9. **Advanced Authentication**: JWT-based token authentication instead of session-based
10. **Role-Based Access Control (RBAC)**: Admin, researcher, and user roles with different permissions
11. **Mobile Application**: Native mobile app for resource management
12. **Email Notifications**: Notification system for resource requests and approvals
13. **Hardware Maintenance Tracking**: Track maintenance windows and hardware downtime
14. **Multi-tenancy**: Support for multiple organizations/tenants on single platform
15. **API Rate Limiting**: Implement rate limiting for API stability
16. **Audit Logging**: Comprehensive audit trail for compliance and troubleshooting

---

## 9. Approval

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Project Manager | TBD | | |
| Development Lead | TBD | | |
| QA Lead | TBD | | |
| Customer Representative | TBD | | |

---

**Document Version History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | Feb 2026 | AI Assistant | Initial SRS creation |

---

## References

1. **ECE 382V: Cloud Native App Development Course Materials**
   - Software Development Life Cycle (SDLC) and Agile Methodology
   - Web UI Development and React.js
   - Python Object-Oriented Programming (OOP)
   - Python Cryptography Modules
   - MongoDB Database Design and Implementation

2. **University of Utah POWDER Program**
   - Reference: https://powderwireless.net/
   - Inspiration for HaaS system architecture and resource management workflows

3. **Course Assignment Document**
   - Team Project: Stakeholder Needs and Grading
   - Version 7.20260116
   - Defines stakeholder needs (SN1-SN6) and system requirements (SR1-SR5)

4. **Technology Documentation**
   - Flask Official Documentation: https://flask.palletsprojects.com/
   - React Official Documentation: https://react.dev/
   - MongoDB Official Documentation: https://docs.mongodb.com/
   - Docker Official Documentation: https://docs.docker.com/
   - Vite Official Documentation: https://vitejs.dev/

5. **Software Requirements Specification Standards**
   - IEEE Std 830-1998 "Recommended Practice for Software Requirements Specifications"
   - ISO/IEC/IEEE 29148:2018 "Systems and software engineering -- Life cycle processes -- Requirements engineering"

6. **Cloud-Native Team Project Repository**
   - Repository: cloud-native-team-project
   - Owner: caseyrwebb
   - Primary Branch: 20-backend-projects-api
   - Default Branch: main

---

*Last Updated: February 10, 2026*
