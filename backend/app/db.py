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
