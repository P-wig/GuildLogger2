import os


class Config:
    MONGO_URI = os.getenv("MONGO_URI", "mongodb://localhost:27017")
    MONGO_DB = os.getenv("MONGO_DB", "myapp")
    MONGO_INIT_COLLECTIONS = [
        name.strip()
        for name in os.getenv("MONGO_INIT_COLLECTIONS", "users").split(",")
        if name.strip()
    ]
