from __future__ import annotations

from flask import current_app
from pymongo import MongoClient
from pymongo.database import Database


def init_mongo(app) -> None:
    app.extensions["mongo_client"] = MongoClient(app.config["MONGO_URI"])


def get_db() -> Database:
    client: MongoClient = current_app.extensions["mongo_client"]
    return client[current_app.config["MONGO_DB"]]


def close_mongo(app) -> None:
    client: MongoClient | None = app.extensions.pop("mongo_client", None)
    if client is not None:
        client.close()
