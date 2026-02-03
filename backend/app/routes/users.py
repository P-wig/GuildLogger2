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
    """
    Docstring for create_user
    """
    data = request.get_json(silent=True) or {}
    # minimal validation (extend later)
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400

    required_fields = ["username", "userid", "password"]
    if not all(field in data for field in required_fields):
        missing_fields = [field for field in required_fields if field not in data]
        error_message = f"Missing fields: {', '.join(missing_fields)}"
        return jsonify({"error": error_message}), 400
    else:
        # Create Hashed Password and UserID
        hash_userid = _encrypt(data["userid"], 3, 1)
        hash_password = _encrypt(data["password"], 3, 1)
        data["userid"] = hash_userid
        data["password"] = hash_password
        res = users_col().insert_one(data)
        doc = users_col().find_one({"_id": res.inserted_id})

        if not doc:
            return jsonify({"error": "Failed to create user"}), 500
        else:
            return jsonify(serialize_doc(sanitize_user(doc))), 201


@bp.get("/<user_id>")
def get_user(user_id: str):
    try:
        _id = to_object_id(user_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    doc = users_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Not found"}), 404
    return jsonify(serialize_doc(sanitize_user(doc)))


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


def sanitize_user(doc: dict) -> dict:
    """Remove sensitive fields from user document."""
    sensitive_fields = ["password", "userid"]
    return {k: v for k, v in doc.items() if k not in sensitive_fields}
