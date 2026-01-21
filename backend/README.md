# Backend Flask Boilerplate

Small Flask + MongoDB API service with an app factory, CORS for local React,
and a simple `init-db` command for first-time Mongo setup.

## Requirements
- Python 3.10+
- MongoDB connection string

## Setup
```bash
python -m venv .venv
source .venv/bin/activate
pip install -e .
```

If you prefer requirements files:
```bash
pip install -r requirements.txt
```

## Configuration
Copy `.env.example` to `.env` and adjust values:
```bash
cp .env.example .env
```

Environment variables:
- `MONGO_URI`: Mongo connection string (default: `mongodb://localhost:27017`)
- `MONGO_DB`: Default database name (default: `myapp`)
- `MONGO_INIT_COLLECTIONS`: Comma-separated collections to create on init

## Run
```bash
python run.py
```

Or use Flask's CLI:
```bash
flask --app app:create_app run --debug
```

## Initialize a database
MongoDB creates a database on first write or collection creation. Use:
```bash
flask --app app:create_app init-db
```

Initialize a different database name:
```bash
flask --app app:create_app init-db --db tenant_foo
```

## Multi-database usage
Use `get_db()` with a database name when needed:
```python
from app.db import get_db

db = get_db("tenant_foo")
db.users.insert_one({"username": "casey"})
```

## API endpoints
- `GET /api/health`
- `GET /api/users`
- `POST /api/users`
