from __future__ import annotations

from typing import Iterable, Optional

import click
from flask import current_app, g
from pymongo import MongoClient
from pymongo.database import Database


def get_client() -> MongoClient:
    client = g.get("mongo_client")
    if client is None:
        client = MongoClient(current_app.config["MONGO_URI"])
        g.mongo_client = client
    return client


def get_db(db_name: Optional[str] = None) -> Database:
    if db_name is None:
        db_name = current_app.config["MONGO_DB"]

    dbs = g.setdefault("mongo_dbs", {})
    if db_name not in dbs:
        dbs[db_name] = get_client()[db_name]

    return dbs[db_name]


def init_db(
    db_name: Optional[str] = None,
    collections: Optional[Iterable[str]] = None,
) -> None:
    db = get_db(db_name=db_name)
    if collections is None:
        collections = current_app.config["MONGO_INIT_COLLECTIONS"]

    if not collections:
        return

    existing = set(db.list_collection_names())
    for name in collections:
        if name not in existing:
            db.create_collection(name)


@click.command("init-db")
@click.option(
    "--db",
    "db_name",
    default=None,
    help="Database name (defaults to MONGO_DB).",
)
def init_db_command(db_name: Optional[str]) -> None:
    init_db(db_name=db_name)
    resolved_name = db_name or current_app.config["MONGO_DB"]
    click.echo(f"Initialized MongoDB database: {resolved_name}")


def close_db(e=None) -> None:
    client = g.pop("mongo_client", None)
    if client is not None:
        client.close()
    g.pop("mongo_dbs", None)


def init_app(app) -> None:
    app.teardown_appcontext(close_db)
    app.cli.add_command(init_db_command)
