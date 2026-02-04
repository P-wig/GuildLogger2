from __future__ import annotations

from flask import Blueprint, jsonify, request
from pymongo.collection import Collection

from app.db import get_db
from app.mongo_utils import serialize_doc, to_object_id

from .auth import _encrypt
bp = Blueprint("users", __name__)


def users_col() -> Collection:
    return get_db()["users"]


@bp.get("")
def list_users():
    docs = [serialize_doc(d) for d in users_col().find().limit(200)]
    return jsonify(docs)


@bp.post("")
def create_user():
    data = request.get_json(silent=True) or {}
    # minimal validation (extend later)
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400

    res = users_col().insert_one(data)
    doc = users_col().find_one({"_id": res.inserted_id})
    if not doc:
        return jsonify({"error": "Failed to create user"}), 500
    return jsonify(serialize_doc(doc)), 201


@bp.get("/<user_id>")
def get_user(user_id: str):
    try:
        _id = to_object_id(user_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    doc = users_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Not found"}), 404
    return jsonify(serialize_doc(doc))


@bp.put("/<user_id>")
def replace_user(user_id: str):
    data = request.get_json(silent=True) or {}
    if not isinstance(data, dict):
        return jsonify({"error": "Expected JSON object body"}), 400
    data.pop("_id", None)

    try:
        _id = to_object_id(user_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = users_col().replace_one({"_id": _id}, data, upsert=False)
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = users_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve user"}), 500
    return jsonify(serialize_doc(doc))


@bp.patch("/<user_id>")
def update_user(user_id: str):
    data = request.get_json(silent=True) or {}
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400
    data.pop("_id", None)

    try:
        _id = to_object_id(user_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = users_col().update_one({"_id": _id}, {"$set": data})
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = users_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve user"}), 500
    return jsonify(serialize_doc(doc))


@bp.delete("/<user_id>")
def delete_user(user_id: str):
    try:
        _id = to_object_id(user_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = users_col().delete_one({"_id": _id})
    if res.deleted_count == 0:
        return jsonify({"error": "Not found"}), 404
    return "", 204
