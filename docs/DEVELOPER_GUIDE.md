# Developer Guide
**Team Project**

Guidelines, standards, and workflows for development team members.

## Table of Contents
1. [Coding Standards](#coding-standards)
2. [Project Structure](#project-structure)
3. [Git Workflow](#git-workflow)
4. [Development Practices](#development-practices)
5. [Code Review Process](#code-review-process)
6. [Testing Requirements](#testing-requirements)
7. [Commit Message Guidelines](#commit-message-guidelines)
8. [Getting Help](#getting-help)

---

## Coding Standards

### Python (Backend)

**Style Guide:** PEP 8

```bash
# Format code with Black
black backend/

# Check style with flake8
flake8 backend/

# Type checking with mypy (recommended)
mypy backend/
```

**Naming Conventions:**
- Functions and variables: `snake_case`
- Classes: `PascalCase`
- Constants: `UPPER_SNAKE_CASE`
- Private variables: prefix with `_`

**Code Organization:**
```python
# Imports (ordered: standard lib, third-party, local)
import os
from datetime import datetime

import flask
from pymongo import MongoClient

from app.routes import auth

# Constants
MAX_USERS = 1000
DEFAULT_TIMEOUT = 30

# Classes
class UserModel:
    pass

# Functions
def create_user(userid: str, password: str) -> dict:
    pass
```

**Documentation:**
```python
def checkout_hardware(project_id: str, hardware_set: str, units: int) -> dict:
    """
    Check out hardware units for a project.
    
    Args:
        project_id: Unique project identifier
        hardware_set: Hardware set name (e.g., 'HWSet1')
        units: Number of units to check out
        
    Returns:
        dict: Allocation record with checkout details
        
    Raises:
        ValueError: If insufficient hardware available
        KeyError: If project_id or hardware_set not found
    """
    pass
```

### TypeScript/React (Frontend)

**Style Guide:** Airbnb TypeScript style guide

```bash
# Format code with Prettier
npm run format

# Lint with ESLint
npm run lint
```

**Naming Conventions:**
- Variables and functions: `camelCase`
- Components and types: `PascalCase`
- Constants: `UPPER_SNAKE_CASE`
- Private/internal: prefix with `_`

**Component Structure:**
```tsx
import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';

interface HardwareCheckoutProps {
  projectId: string;
  onSuccess?: () => void;
}

/**
 * Hardware checkout form component
 * @param props Component props
 * @returns React component
 */
export const HardwareCheckout: React.FC<HardwareCheckoutProps> = ({ 
  projectId, 
  onSuccess 
}) => {
  const [units, setUnits] = useState<number>(1);
  const navigate = useNavigate();

  return (
    <Box sx={{ p: 2 }}>
      {/* Component JSX */}
    </Box>
  );
};
```

### General Principles

1. **DRY (Don't Repeat Yourself)** - Extract reusable functions and components
2. **KISS (Keep It Simple, Stupid)** - Clear, simple code over clever code
3. **Single Responsibility** - Each function/component has one job
4. **Error Handling** - Handle errors explicitly, don't silently fail
5. **Documentation** - Comment why, not what (code shows what)

---

## Project Structure

### Backend
```
backend/
├── app/
│   ├── __init__.py              # Flask app factory
│   ├── config.py                # Configuration
│   ├── db.py                    # Database connection
│   ├── routes/
│   │   ├── auth.py              # Authentication endpoints
│   │   ├── users.py             # User management endpoints
│   │   ├── projects.py          # Project endpoints
│   │   ├── hardware.py          # Hardware checkout endpoints
│   │   └── health.py            # Health check
│   └── mongo_utils.py           # MongoDB utility functions
├── pyproject.toml               # Project metadata and dependencies
├── run.py                       # Application entry point
└── Dockerfile                   # Container configuration
```

**Creating New Routes:**
1. Create new file in `routes/` folder
2. Define Flask blueprint
3. Register blueprint in `app/__init__.py`

```python
# routes/new_feature.py
from flask import Blueprint

bp = Blueprint('new_feature', __name__, url_prefix='/api')

@bp.route('/new-endpoint', methods=['GET'])
def get_new_endpoint():
    return {'status': 'ok'}
```

### Frontend
```
frontend/
├── src/
│   ├── main.tsx                 # Application entry point
│   ├── App.tsx                  # Root component
│   ├── index.css                # Global styles
│   ├── pages/
│   │   ├── Home.tsx             # Home page
│   │   ├── Auth.tsx             # Login/register
│   │   ├── Account.tsx          # User profile
│   │   ├── Projects.tsx         # Project management (TODO)
│   │   └── Hardware.tsx         # Hardware checkout (TODO)
│   ├── components/              # Reusable components
│   ├── api/
│   │   ├── http.ts              # HTTP client wrapper
│   │   ├── users.ts             # User API functions
│   │   ├── projects.ts          # Project API functions
│   │   └── hardware.ts          # Hardware API functions (TODO)
│   ├── auth/
│   │   ├── authContext.ts       # Auth context/state
│   │   ├── AuthProvider.tsx     # Auth provider component
│   │   └── useAuth.ts           # Auth hook
│   ├── routes/
│   │   ├── index.ts             # Route definitions
│   │   ├── router.tsx           # Router configuration
│   │   └── ProtectedRoute.tsx   # Route guard component
│   ├── layouts/
│   │   ├── AppLayout.tsx        # Main layout wrapper
│   │   └── index.ts
│   ├── types/                   # TypeScript type definitions
│   └── styles/                  # Component-specific styles
├── public/                      # Static assets
├── index.html                   # HTML entry point
├── vite.config.ts              # Vite configuration
├── tsconfig.json               # TypeScript configuration
├── eslint.config.js            # ESLint configuration
└── Dockerfile                  # Container configuration
```

---

## Git Workflow

### Branch Naming

```
feature/description         # New feature (e.g., feature/hardware-checkout)
bugfix/issue-name          # Bug fix (e.g., bugfix/auth-token-validation)
docs/update-api-docs       # Documentation (e.g., docs/api-reference)
refactor/improve-component # Refactoring (e.g., refactor/auth-context)
chore/update-deps          # Maintenance (e.g., chore/bump-react-version)
```

### Workflow Steps

1. **Create feature branch from main/develop:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and commit regularly:**
   ```bash
   git add .
   git commit -m "type: description"
   ```

3. **Push to remote:**
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create Pull Request:**
   - Clear description of changes
   - Link to related issues
   - Screenshots for UI changes
   - Test results

5. **Code review and merge:**
   - Minimum 1 approval required
   - All CI checks must pass
   - Squash commits before merging

### Protecting Main Branch

- Require pull request reviews
- Require status checks to pass before merging
- Require branches to be up to date before merging
- Require code review from code owners (if applicable)

---

## Development Practices

### Before Coding

1. Check existing issues/PRs to avoid duplicating work
2. Create or reference a GitHub issue
3. Plan your changes before coding
4. Check the design system requirements

### While Coding

1. **Small, focused commits:**
   ```bash
   git commit -m "feature: add checkout validation"
   git commit -m "refactor: simplify allocation logic"
   ```

2. **Test as you go:**
   - Run linter: `npm run lint` (frontend) / `black --check` (backend)
   - Run tests: `npm test` (frontend) / `pytest` (backend)
   - Manual testing in dev environment

3. **Keep PRs reviewable:**
   - Max 400 lines of code per PR (if possible)
   - Related changes together, unrelated changes in separate PRs
   - Avoid formatting-only changes mixed with logic changes

### Before Requesting Review

1. **Backend:**
   ```bash
   black backend/           # Format code
   flake8 backend/         # Check style
   pytest backend/         # Run tests (TODO)
   ```

2. **Frontend:**
   ```bash
   npm run format          # Format code
   npm run lint            # Check style
   npm test                # Run tests (TODO)
   npm run build           # Verify build works
   ```

3. **Manual Testing:**
   - Test new features locally
   - Test error cases
   - Test edge cases
   - Verify no console errors

---

## Code Review Process

### As an Author

1. **Clear PR description:**
   ```markdown
   ## Changes
   - What was changed and why
   
   ## Testing
   - How to test these changes
   
   ## Screenshots/Videos
   - If UI changes, include screenshots
   
   ## Checklist
   - [ ] Tests pass
   - [ ] Code follows style guide
   - [ ] Documentation updated
   - [ ] No breaking changes
   ```

2. **Respond to feedback promptly** (within 24 hours)
3. **Request re-review** after making changes

### As a Reviewer

1. **Review for:**
   - Code quality and style consistency
   - Logic and correctness
   - Test coverage
   - Documentation
   - Performance concerns

2. **Be constructive:**
   - Ask questions instead of making demands
   - Suggest improvements with examples
   - Acknowledge good solutions

3. **Approve when:**
   - Code is correct and clear
   - Tests are adequate
   - No unresolved comments

---

## Testing Requirements

**TODO:** Establish testing thresholds and requirements

Current guidance:
- Aim for 80%+ code coverage
- Test happy path and error cases
- Test edge cases and boundary conditions
- Integration tests for API endpoints

See [TESTING.md](./TESTING.md) for detailed testing guidance.

---

## Commit Message Guidelines

### Format
```
type(scope): subject

body

footer
```

### Types
- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation changes
- **style:** Code style changes (formatting, semicolons, etc.)
- **refactor:** Code refactoring without feature changes
- **test:** Adding or updating tests
- **chore:** Dependency updates, build config, etc.

### Examples
```
feat(auth): add password reset functionality

Users can now reset their forgotten passwords via email.
Implemented token-based verification with 24-hour expiry.

Closes #123
```

```
fix(hardware): prevent overallocation with atomic updates

Use MongoDB $inc operator to prevent race conditions
during concurrent hardware checkout operations.

Fixes #456
```

```
docs: update API reference for hardware endpoints
```

---

## Getting Help

### Documentation
- [README.md](../README.md) - Setup and running
- [API_REFERENCE.md](./API_REFERENCE.md) - API endpoints
- [DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md) - UI/Design standards
- [CONFIGURATION.md](./CONFIGURATION.md) - Environment variables
- [TESTING.md](./TESTING.md) - Testing strategy

### Communication
- **Questions:** Ask in team chat or open a discussion issue
- **Bugs:** Create a GitHub issue with reproduction steps
- **Design decisions:** Discuss in PR comments or team meetings

### Useful Links
- [Flask Documentation](https://flask.palletsprojects.com/)
- [React Documentation](https://react.dev/)
- [Material-UI Documentation](https://mui.com/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [Git Documentation](https://git-scm.com/doc)

---

**Last Updated:** February 13, 2026  
**Version:** 1.1  
**Status:** Active
