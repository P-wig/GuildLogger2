import os


class Config:
    # Mongo
    MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
    MONGO_DB = os.getenv("MONGO_DB", "test")

    # CORS (Vite default)
    # Comma-separated list in .env, e.g. "http://localhost:5173,http://127.0.0.1:5173"
    CORS_ORIGINS = [
        o.strip()
        for o in os.getenv("CORS_ORIGINS", "http://localhost:5173").split(",")
        if o.strip()
    ]
