# Requirements Compliance and Code Changes

**Team Project - Hardware Checkout System**  
**Version:** 1.0  
**Date:** February 13, 2026

This document tracks requirements compliance and lists all code changes needed.

---

## Table of Contents
1. [Requirements Not Met](#requirements-not-met)
2. [Code Changes Required](#code-changes-required)
3. [Implementation Priority](#implementation-priority)

---

## Requirements Not Met

### SR3: Encrypt User-ID and Password (NOT MET)

**Requirement:** "PoC App shall have a mechanism for encrypting user-id and password"

**Current State:** Only password is encrypted. userId is stored and compared in plaintext.

**Evidence (auth.py lines 67-70):**
```python
# Current code - WRONG
user = db["users"].find_one({"userId": userid, "password": encrypt_password})
```

**Fix Required:**
```python
# Correct code - encrypt BOTH
encrypt_userid = _encrypt(userid, 3, 1)
encrypt_password = _encrypt(password, 3, 1)
user = db["users"].find_one({"userId": encrypt_userid, "password": encrypt_password})
```

**Additional Fix for Display:**
Add decrypt function to retrieve userId for display purposes:
```python
def _decrypt(encrypted_text: str, num_shift: int, dir_shift: int) -> str:
    """Decrypt by reversing the cipher direction."""
    return _encrypt(encrypted_text, num_shift, -dir_shift)
```

**Files to Change:**
| File | Change |
|------|--------|
| backend/app/routes/auth.py | Encrypt userId in login() and register() |
| backend/app/routes/auth.py | Add _decrypt() function |
| backend/app/routes/auth.py | Return decrypted userId in response |

---

### SR5: Database for User/Project/Resource Details (PARTIALLY MET)

**Requirement:** "PoC App shall have a database for maintaining user login credentials, project codes, project details, resource details"

**Current State:** 
- Users collection: EXISTS
- Projects collection: EXISTS (basic)
- Hardware collection: NEEDS BUILD
- Allocations collection: NEEDS BUILD

**Fix Required:** Create hardware and allocations collections with proper schema.

---

### Feature 6 (User Management): Security Features (NOT MET)

**Requirement:** "Security features to encrypt the userid and password"

**Current State:** 
- Password encrypted: YES
- UserId encrypted: NO
- Console.log exposes credentials: YES (security vulnerability)

**Files with Security Issues:**
| File | Issue | Line |
|------|-------|------|
| frontend/src/pages/Account.tsx | console.log prints user data | ~10 |
| frontend/src/pages/Auth.tsx | May log credentials in error handling | Various |

---

### Feature 1-5 (Resource Management): Hardware Checkout (NOT MET)

**Requirements:**
1. Display area showing capacity of HWSet1 and HWSet2
2. Display area showing availability of HWSet1 and HWSet2
3. Database where HW information can be stored and retrieved
4. Display area for checkout/check-in quantities

**Current State:** Hardware endpoints not implemented.

---

## Code Changes Required

### 1. Backend: Fix userId Encryption (auth.py)

**File:** `backend/app/routes/auth.py`

**Change 1: Add decrypt function (after _encrypt function)**
```python
def _decrypt(encrypted_text: str, num_shift: int, dir_shift: int) -> str:
    """
    Decrypt using the cyclic cipher algorithm (reverse direction).
    
    :param encrypted_text: text to decrypt
    :param num_shift: number of shifts (same as encryption)
    :param dir_shift: direction shift (same as encryption, will be reversed)
    :return: decrypted string
    """
    return _encrypt(encrypted_text, num_shift, -dir_shift)
```

**Change 2: Fix login() function**
```python
@bp.post("/login")
def login():
    data = request.get_json()
    userid = data.get("userId")
    password = data.get("password")

    if userid is None or password is None:
        return jsonify({"ok": False, "error": "Missing userId or password"}), 400

    # Encrypt BOTH userId and password for comparison
    encrypt_userid = _encrypt(userid, 3, 1)
    encrypt_password = _encrypt(password, 3, 1)

    db = get_db()
    user = db["users"].find_one({
        "userId": encrypt_userid,
        "password": encrypt_password
    })

    if user is None:
        return jsonify({"ok": False, "error": "Invalid credentials"}), 401

    # Return UNENCRYPTED userId for client display
    return jsonify({
        "ok": True,
        "message": "Login successful",
        "user": {"userId": userid}  # Original unencrypted value
    }), 200
```

**Change 3: Fix register() function**
```python
@bp.post("/register")
def register():
    data = request.get_json()
    required_fields = ["userId", "password"]  # Only these two per professor
    missing_fields = [field for field in required_fields if field not in data]

    if missing_fields:
        error_message = f"Missing fields: {', '.join(missing_fields)}"
        return jsonify({"error": error_message}), 400

    # Store original userId for response
    original_userid = data["userId"]
    
    # Encrypt userId for storage and uniqueness check
    encrypt_userid = _encrypt(data["userId"], 3, 1)
    
    # Check if encrypted userId already exists
    if users_col().find_one({"userId": encrypt_userid}):
        return jsonify({"error": "userId already exists"}), 409

    # Encrypt password
    encrypt_password = _encrypt(data["password"], 3, 1)
    
    # Store encrypted values
    data["userId"] = encrypt_userid
    data["password"] = encrypt_password
    
    res = users_col().insert_one(data)
    doc = users_col().find_one({"_id": res.inserted_id})

    if not doc:
        return jsonify({"error": "Failed to create user"}), 500

    # Return UNENCRYPTED userId for client use
    return jsonify({"user": {"userId": original_userid}}), 201
```

---

### 2. Frontend: Remove Console.log Security Vulnerabilities

**File:** `frontend/src/pages/Account.tsx`

**Remove this line:**
```typescript
// DELETE THIS LINE
console.log("User data in Account component:", user);
```

**File:** `frontend/src/pages/Auth.tsx`

**Review and remove any credential logging in error handling.**

---

### 3. Backend: Create Hardware Endpoints (NEW FILE)

**File:** `backend/app/routes/hardware.py`

```python
"""Hardware management routes for checkout/checkin functionality."""
from flask import Blueprint, jsonify, request
from app.db import get_db
from app.mongo_utils import serialize_doc

bp = Blueprint("hardware", __name__)

def hardware_col():
    """Get hardware collection."""
    return get_db()["hardware"]

def allocations_col():
    """Get allocations collection."""
    return get_db()["allocations"]


@bp.get("/")
def get_all_hardware():
    """Get all hardware sets with availability."""
    hardware_list = list(hardware_col().find())
    return jsonify([serialize_doc(hw) for hw in hardware_list]), 200


@bp.get("/availability")
def get_availability():
    """Get real-time availability of all hardware."""
    hardware_list = list(hardware_col().find())
    result = []
    for hw in hardware_list:
        capacity = hw.get("capacity", 0)
        available = hw.get("available", 0)
        result.append({
            "hardwareId": hw.get("hardwareId"),
            "hardwareName": hw.get("hardwareName"),
            "capacity": capacity,
            "available": available,
            "percentageAvailable": (available / capacity * 100) if capacity > 0 else 0
        })
    return jsonify(result), 200


@bp.post("/<hardware_id>/checkout")
def checkout_hardware(hardware_id):
    """Check out hardware units for a project."""
    data = request.get_json()
    project_id = data.get("projectId")
    qty = data.get("qty", 0)
    user_id = data.get("userId")

    if not all([project_id, qty > 0, user_id]):
        return jsonify({"error": "Missing projectId, qty, or userId"}), 400

    # Find hardware
    hardware = hardware_col().find_one({"hardwareId": hardware_id})
    if not hardware:
        return jsonify({"error": "Hardware not found"}), 404

    # Check availability
    available = hardware.get("available", 0)
    if qty > available:
        return jsonify({
            "error": "Insufficient hardware available",
            "available": available,
            "requested": qty
        }), 409

    # Atomic update - decrement availability
    result = hardware_col().find_one_and_update(
        {"hardwareId": hardware_id, "available": {"$gte": qty}},
        {"$inc": {"available": -qty}},
        return_document=True
    )

    if not result:
        return jsonify({"error": "Checkout failed - concurrent modification"}), 409

    # Record allocation
    from datetime import datetime
    allocation = {
        "projectId": project_id,
        "hardwareId": hardware_id,
        "userId": user_id,
        "qty": qty,
        "type": "checkout",
        "timestamp": datetime.utcnow().isoformat()
    }
    allocations_col().insert_one(allocation)

    return jsonify({
        "projectId": project_id,
        "hardwareId": hardware_id,
        "qty": qty,
        "type": "checkout",
        "available": result.get("available")
    }), 200


@bp.post("/<hardware_id>/checkin")
def checkin_hardware(hardware_id):
    """Check in hardware units from a project."""
    data = request.get_json()
    project_id = data.get("projectId")
    qty = data.get("qty", 0)
    user_id = data.get("userId")

    if not all([project_id, qty > 0, user_id]):
        return jsonify({"error": "Missing projectId, qty, or userId"}), 400

    # Find hardware
    hardware = hardware_col().find_one({"hardwareId": hardware_id})
    if not hardware:
        return jsonify({"error": "Hardware not found"}), 404

    # Atomic update - increment availability
    capacity = hardware.get("capacity", 0)
    result = hardware_col().find_one_and_update(
        {"hardwareId": hardware_id},
        {"$inc": {"available": qty}},
        return_document=True
    )

    # Cap at capacity
    if result and result.get("available") > capacity:
        hardware_col().update_one(
            {"hardwareId": hardware_id},
            {"$set": {"available": capacity}}
        )
        result["available"] = capacity

    # Record allocation
    from datetime import datetime
    allocation = {
        "projectId": project_id,
        "hardwareId": hardware_id,
        "userId": user_id,
        "qty": qty,
        "type": "checkin",
        "timestamp": datetime.utcnow().isoformat()
    }
    allocations_col().insert_one(allocation)

    return jsonify({
        "projectId": project_id,
        "hardwareId": hardware_id,
        "qty": qty,
        "type": "checkin",
        "available": result.get("available")
    }), 200
```

**File:** `backend/app/__init__.py`

**Add hardware blueprint registration:**
```python
from app.routes import hardware
app.register_blueprint(hardware.bp, url_prefix="/api/hardware")
```

---

### 4. Database: Initialize Hardware Data

**Create script or seed data for HWSet1 and HWSet2:**

```python
# backend/scripts/init_hardware.py
from app.db import get_db

def init_hardware():
    db = get_db()
    hardware_col = db["hardware"]
    
    # Clear existing
    hardware_col.delete_many({})
    
    # Insert HWSet1 and HWSet2
    hardware_col.insert_many([
        {
            "hardwareId": "HWSet1",
            "hardwareName": "Hardware Set 1",
            "capacity": 100,
            "available": 100
        },
        {
            "hardwareId": "HWSet2", 
            "hardwareName": "Hardware Set 2",
            "capacity": 100,
            "available": 100
        }
    ])
    print("Hardware initialized successfully")

if __name__ == "__main__":
    init_hardware()
```

---

### 5. Frontend: Projects Page Integration

**File:** `frontend/src/pages/Projects.tsx` (Anita's work)

Needs to integrate with:
- GET /api/projects - fetch user's projects
- GET /api/hardware/availability - show hardware status
- POST /api/hardware/{id}/checkout - checkout hardware
- POST /api/hardware/{id}/checkin - checkin hardware

---

## Implementation Priority

| Priority | Task | Owner | Requirement |
|----------|------|-------|-------------|
| HIGH | Fix userId encryption in auth.py | Yuri/Casey | SR3 |
| HIGH | Remove console.log security issues | Isaac | SR3 |
| HIGH | Create hardware endpoints | Yuri | SR5, Features 1-4 |
| HIGH | Initialize HWSet1/HWSet2 data | Yuri | Features 1-2 |
| MEDIUM | Projects page API integration | Anita | SR4 |
| MEDIUM | Hardware checkout UI | Anita | Feature 4 |
| LOW | Add username/display name field | Team | Nice-to-have |

---

## Summary Checklist

- [ ] Encrypt userId in login() - auth.py
- [ ] Encrypt userId in register() - auth.py  
- [ ] Add _decrypt() function - auth.py
- [ ] Remove console.log in Account.tsx
- [ ] Review Auth.tsx for credential logging
- [ ] Create hardware.py with GET/POST endpoints
- [ ] Register hardware blueprint in __init__.py
- [ ] Create init_hardware.py script
- [ ] Run script to create HWSet1 and HWSet2
- [ ] Integrate Projects page with backend APIs
- [ ] Build hardware checkout/checkin UI

---

**Last Updated:** February 13, 2026
