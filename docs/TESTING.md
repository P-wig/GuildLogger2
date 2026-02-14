# Testing Documentation
**Team Project**

Comprehensive testing strategy, guidelines, and implementation details.

## Table of Contents
1. [Testing Strategy](#testing-strategy)
2. [Test Types](#test-types)
3. [Backend Testing](#backend-testing)
4. [Frontend Testing](#frontend-testing)
5. [Integration Testing](#integration-testing)
6. [Running Tests](#running-tests)
7. [Coverage Goals](#coverage-goals)
8. [Best Practices](#best-practices)

---

## Testing Strategy

### Overview
We employ a **pyramid testing approach**:
- **Unit Tests** (base) - smallest units of code in isolation
- **Integration Tests** (middle) - multiple components working together
- **End-to-End Tests** (top) - complete user workflows

### Goals
- Catch bugs early in development
- Prevent regressions when refactoring
- Document expected behavior
- Enable confident refactoring
- Improve code quality

### Test Environments
- **Development:** Local testing during development
- **Staging:** Automated tests on commits (CI/CD)
- **Production:** Smoke tests for critical paths

---

## Test Types

### Unit Tests
**What:** Test individual functions/methods in isolation

**When:** Always - for every new function

**Example (Python):**
```python
import pytest
from app.routes.users import sanitize_user

def test_sanitize_user_removes_password():
    """Verify password is removed from user object."""
    user = {
        "userId": "john_doe",
        "password": "encrypted_pw_123"
    }
    result = sanitize_user(user)
    assert "password" not in result
    assert result["userId"] == "john_doe"
```

**Example (TypeScript):**
```typescript
import { validateEmail } from '@/api/validators';

describe('validateEmail', () => {
  it('should accept valid email addresses', () => {
    expect(validateEmail('user@example.com')).toBe(true);
  });

  it('should reject invalid email addresses', () => {
    expect(validateEmail('not-an-email')).toBe(false);
  });
});
```

### Integration Tests
**What:** Test multiple components working together

**When:** For API endpoints, database operations, complex workflows

**Example (Python/Flask):**
```python
def test_create_user_stores_in_database(client, db):
    """Verify user creation endpoint stores data in MongoDB."""
    response = client.post('/api/users/register', json={
        "userId": "test_user",
        "password": "securepass123"
    })
    assert response.status_code == 201
    
    # Verify data in database (userId stored encrypted)
    # Note: query by encrypted userId in production
    assert response.json()["userId"] == "test_user"  # Returned decrypted
```

### End-to-End Tests
**What:** Test complete user workflows through UI

**When:** For critical paths (login, checkout, etc.)

**Example (Playwright/Cypress - Future):**
```javascript
test('user can login and checkout hardware', async ({ page }) => {
  // Navigate to login
  await page.goto('/auth');
  
  // Login
  await page.fill('[name="userid"]', 'testuser');
  await page.fill('[name="password"]', 'password123');
  await page.click('button:has-text("Sign In")');
  
  // Wait for redirect
  await page.waitForURL('/account');
  
  // Navigate to hardware
  await page.click('a:has-text("Hardware")');
  
  // Checkout hardware
  await page.click('button:has-text("Request Checkout")');
  await page.fill('[name="units"]', '2');
  await page.click('button:has-text("Checkout")');
  
  // Verify success message
  await expect(page.locator('text=Checkout successful')).toBeVisible();
});
```

---

## Backend Testing

### Framework
**Current:** TODO (needs to be established)

**Recommendations:**
- **pytest** - Python testing framework
- **pytest-flask** - Flask fixtures
- **pytest-mongodb** - MongoDB fixtures
- **coverage** - Code coverage measurement

### Setup

```bash
# Install testing dependencies (when established)
pip install pytest pytest-flask pytest-mongodb coverage pytest-cov

# Create test file structure
backend/
├── tests/
│   ├── __init__.py
│   ├── conftest.py              # Shared fixtures
│   ├── test_auth.py
│   ├── test_users.py
│   ├── test_projects.py
│   ├── test_hardware.py
│   └── unit/
│       └── test_mongo_utils.py
```

### Test Fixtures (conftest.py)

```python
import pytest
from app import create_app
from pymongo import MongoClient

@pytest.fixture
def app():
    """Create application for testing."""
    app = create_app(testing=True)
    yield app

@pytest.fixture
def client(app):
    """Test client."""
    return app.test_client()

@pytest.fixture
def db(app):
    """Test database connection."""
    client = MongoClient(app.config['MONGODB_URI'])
    db = client['test_db']
    yield db
    # Cleanup
    client.drop_database('test_db')

@pytest.fixture
def auth_token(client, db):
    """Create test user and return auth token."""
    # TODO: Create test user and generate token
    pass
```

### Writing Backend Tests

```python
# tests/test_users.py
import pytest

class TestUserCreation:
    """User creation endpoint tests."""
    
    def test_create_user_success(self, client):
        """Test successful user creation."""
        response = client.post('/api/users/register', json={
            "userId": "newuser",
            "password": "secure123"
        })
        
        assert response.status_code == 201
        data = response.get_json()
        assert data['userId'] == "newuser"  # Returned decrypted for display
        assert 'password' not in data  # Never return password
    
    def test_create_user_duplicate_userid(self, client, db):
        """Test creation fails with duplicate userId."""
        # Create first user
        client.post('/api/users/register', json={
            "userId": "duplicate",
            "password": "pass123"
        })
        
        # Try to create duplicate
        response = client.post('/api/users/register', json={
            "userId": "duplicate",
            "password": "pass456"
        })
        
        assert response.status_code == 409
        assert response.get_json()['error'] == "User already exists"
    
    def test_create_user_missing_fields(self, client):
        """Test creation fails with missing required fields."""
        response = client.post('/api/users/register', json={
            "userId": "newuser"
            # Missing password
        })
        
        assert response.status_code == 400
        assert "required" in response.get_json()['error'].lower()

class TestHardwareCheckout:
    """Hardware checkout endpoint tests."""
    
    def test_checkout_hardware_success(self, client, db, auth_token):
        """Test successful hardware checkout."""
        # TODO: Setup test data
        response = client.post('/api/hardware/checkout',
            headers={'Authorization': f'Bearer {auth_token}'},
            json={
                "projectId": "test_project",
                "hardwareSet": "HWSet1",
                "units": 2
            }
        )
        
        assert response.status_code == 200
        data = response.get_json()
        assert data['units'] == 2
    
    def test_checkout_prevents_overallocation(self, client, db, auth_token):
        """Test checkout fails when insufficient units available."""
        # TODO: Setup limited hardware
        response = client.post('/api/hardware/checkout',
            headers={'Authorization': f'Bearer {auth_token}'},
            json={
                "projectId": "test_project",
                "hardwareSet": "HWSet1",
                "units": 100  # More than available
            }
        )
        
        assert response.status_code == 409
        assert "Insufficient" in response.get_json()['error']
```

---

## Frontend Testing

### Framework
**Current:** TODO (needs to be established)

**Recommendations:**
- **Vitest** - Fast unit testing (works with Vite)
- **React Testing Library** - Component testing
- **@testing-library/user-event** - User interaction simulation
- **MSW (Mock Service Worker)** - API mocking

### Setup

```bash
# Install testing dependencies (when established)
npm install --save-dev vitest @vitest/ui
npm install --save-dev @testing-library/react @testing-library/jest-dom
npm install --save-dev @testing-library/user-event
npm install --save-dev msw

# Create test file structure
frontend/src/
├── __tests__/
│   ├── setup.ts               # Test configuration
│   ├── mocks/
│   │   └── handlers.ts        # API mocks
│   ├── pages/
│   │   ├── Auth.test.tsx
│   │   ├── Home.test.tsx
│   │   └── Hardware.test.tsx
│   ├── components/
│   │   └── LoginForm.test.tsx
│   └── api/
│       └── users.test.ts
```

### Test Setup (setup.ts)

```typescript
import { expect, afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import '@testing-library/jest-dom';
import { server } from './mocks/handlers';

// Start MSW mock server
beforeAll(() => server.listen());
afterEach(() => {
  cleanup();
  server.resetHandlers();
});
afterAll(() => server.close());
```

### API Mocks (mocks/handlers.ts)

```typescript
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';

export const handlers = [
  // Mock GET /api/users
  http.get('/api/users', () => {
    return HttpResponse.json([
      { userId: 'user1' },
      { userId: 'user2' }
    ]);
  }),
  
  // Mock POST /api/auth/login
  http.post('/api/auth/login', async ({ request }) => {
    const body = await request.json();
    
    if (body.userId === 'testuser' && body.password === 'password123') {
      return HttpResponse.json({
        user: { userId: 'testuser' }
      });
    }
    
    return HttpResponse.json(
      { error: 'Invalid credentials' },
      { status: 401 }
    );
  }),
];

export const server = setupServer(...handlers);
```

### Writing Frontend Tests

```typescript
// src/__tests__/pages/Auth.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Auth } from '@/pages/Auth';
import { AuthProvider } from '@/auth/AuthProvider';

describe('Auth Page', () => {
  it('renders login form by default', () => {
    render(
      <AuthProvider>
        <Auth />
      </AuthProvider>
    );
    
    expect(screen.getByText('Sign In')).toBeInTheDocument();
    expect(screen.getByLabelText(/userid/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it('handles successful login', async () => {
    const user = userEvent.setup();
    render(
      <AuthProvider>
        <Auth />
      </AuthProvider>
    );
    
    // Fill form
    await user.type(screen.getByLabelText(/userid/i), 'testuser');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    
    // Submit
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    // Verify success state
    await waitFor(() => {
      expect(screen.queryByLabelText(/userid/i)).not.toBeInTheDocument();
    });
  });

  it('displays error on login failure', async () => {
    const user = userEvent.setup();
    render(
      <AuthProvider>
        <Auth />
      </AuthProvider>
    );
    
    // Fill with invalid credentials
    await user.type(screen.getByLabelText(/userid/i), 'wronguser');
    await user.type(screen.getByLabelText(/password/i), 'wrongpass');
    
    // Submit
    await user.click(screen.getByRole('button', { name: /sign in/i }));
    
    // Verify error message
    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
    });
  });
});
```

---

## Integration Testing

### API Integration Tests

Test that frontend and backend communicate correctly:

```python
# tests/test_integration_auth.py
def test_full_login_flow(client, db):
    """Test complete login workflow."""
    # 1. Create user via registration
    register_response = client.post('/api/users/register', json={
        "userId": "testuser",
        "password": "password123"
    })
    assert register_response.status_code == 201
    
    # 2. Login with created user
    login_response = client.post('/api/auth/login', json={
        "userId": "testuser",
        "password": "password123"
    })
    assert login_response.status_code == 200
    
    token = login_response.get_json()['token']
    
    # 3. Access protected endpoints with token
    projects_response = client.get(
        '/api/projects',
        headers={'Authorization': f'Bearer {token}'}
    )
    assert projects_response.status_code == 200
```

---

## Running Tests

### Backend Tests

```bash
# TODO: Update when testing framework established

# Run all tests
pytest

# Run specific test file
pytest tests/test_users.py

# Run specific test
pytest tests/test_users.py::TestUserCreation::test_create_user_success

# Run with coverage
pytest --cov=app --cov-report=html

# Run with verbose output
pytest -v

# Run with markers
pytest -m "not slow"
```

### Frontend Tests

```bash
# TODO: Update when testing framework established

# Run all tests
npm test

# Run in watch mode (useful during development)
npm test -- --watch

# Run with coverage
npm test -- --coverage

# Run specific test file
npm test -- Auth.test.tsx

# Run UI for test explorer
npm test -- --ui
```

---

## Coverage Goals

**Target:** 80% overall code coverage

| Area | Target | Priority |
|------|--------|----------|
| API endpoints | 90% | HIGH |
| Validation logic | 100% | HIGH |
| Authentication | 90% | HIGH |
| Hardware allocation | 95% | HIGH |
| UI components | 70% | MEDIUM |
| Utility functions | 85% | MEDIUM |
| Error handling | 85% | MEDIUM |

---

## Best Practices

### Do's

1. **Test behavior, not implementation**
   ```python
   # Good - tests the behavior
   def test_password_is_hashed():
       user = create_user("test", "plaintext_password")
       assert user['password_hash'] != "plaintext_password"
   
   # Bad - tests the implementation
   def test_uses_bcrypt():
       assert 'bcrypt' in imports
   ```

2. **Use descriptive test names**
   ```python
   # Good
   def test_checkout_prevents_overallocation_with_insufficient_units():
       pass
   
   # Bad
   def test_checkout():
       pass
   ```

3. **One assertion per test (when possible)**
   ```python
   # Better - focused tests
   def test_user_creation_sets_userid():
       user = create_user("id123", password="test123")
       assert user['userId'] == "id123"
   
   def test_user_creation_encrypts_password():
       user = create_user("id123", password="plain")
       assert user['password'] != "plain"  # Encrypted
   
   # Less ideal - multiple assertions
   def test_user_creation():
       user = create_user("id123", "test123")
       assert user['userId'] == "id123"
       assert user['password'] != "plain"
       assert user['createdAt'] is not None
   ```

4. **Use fixtures for common setup**
   ```python
   # Good - DRY
   @pytest.fixture
   def test_user(db):
       return db['users'].insert_one({...})
   
   def test_something(test_user):
       # test_user fixture provides data
   ```

5. **Test edge cases and error conditions**
   ```python
   # Good - comprehensive
   def test_negative_units_rejected():
       pass
   
   def test_zero_units_rejected():
       pass
   
   def test_max_capacity_enforced():
       pass
   ```

### Don'ts

1. **Don't test the testing framework**
   ```python
   # Pointless
   def test_assert_works():
       assert 1 == 1
   ```

2. **Don't share state between tests**
   ```python
   # Bad - tests affect each other
   test_counter = 0
   
   def test_first():
       global test_counter
       test_counter += 1
   
   # Bad - test order matters
   def test_second():
       assert test_counter == 1  # Fails if test_first didn't run
   ```

3. **Don't make tests too brittle**
   ```python
   # Bad - breaks on implementation changes
   def test_user_uses_database():
       with patch('pymongo.MongoClient') as mock:
           create_user(...)
           mock.assert_called_once()
   
   # Good - tests the behavior
   def test_user_persisted(db):
       create_user(...)
       user = db['users'].find_one()
       assert user is not None
   ```

4. **Don't ignore test failures**
   - Fix immediately, don't skip with `@skip`
   - If skipping is temporary, use `@pytest.mark.skip(reason="TODO: implement")`

---

## CI/CD Integration

**TODO:** Set up continuous integration

Recommended:
- Run tests on every push
- Block merges if tests fail
- Generate coverage reports
- Run linters before tests

---

**Last Updated:** February 13, 2026  
**Version:** 1.1  
**Status:** In Development (Framework Selection Pending)
