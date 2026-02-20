from __future__ import annotations

from flask import current_app
from pymongo import MongoClient
from pymongo.database import Database
from pymongo.errors import ConnectionFailure


def init_mongo(app) -> None:
    try:
        client = MongoClient(app.config["MONGO_URI"], serverSelectionTimeoutMS=5000)
        # Test connection
        client.admin.command("ping")
        app.extensions["mongo_client"] = client
        app.logger.info("MongoDB connected successfully")
    except ConnectionFailure as e:
        app.logger.error(f"MongoDB connection failed: {e}")
        raise


def get_db() -> Database:
    client: MongoClient = current_app.extensions["mongo_client"]
    return client[current_app.config["MONGO_DB"]]


def close_mongo(app) -> None:
    client: MongoClient | None = app.extensions.get("mongo_client", None)
    if client is not None:
        client.close()
        app.extensions.pop("mongo_client", None)


INITIAL_HARDWARE = [
    {
        "hardwareName": "HWSet1",
        "capacity": 200,
        "available": 200,
        "assignedProjects": [],
    },
    {
        "hardwareName": "HWSet2",
        "capacity": 200,
        "available": 200,
        "assignedProjects": [],
    },
    {
        "hardwareName": "HWSet3",
        "capacity": 100,
        "available": 100,
        "assignedProjects": [],
    },
    {
        "hardwareName": "HWSet4",
        "capacity": 100,
        "available": 100,
        "assignedProjects": [],
    },
]


def seed_hardware(db: Database) -> None:
    """Insert default hardware sets if they don't already exist."""
    col = db["hardware"]
    for hw in INITIAL_HARDWARE:
        if not col.find_one({"hardwareName": hw["hardwareName"]}):
            col.insert_one(hw.copy())
