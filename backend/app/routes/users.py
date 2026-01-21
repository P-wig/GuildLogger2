from flask import Blueprint, jsonify, request
from app.db import get_db

def get_blueprint():
    bp = Blueprint("users", __name__, url_prefix="/api/users")

    @bp.get("")
    def list_users():
        db = get_db()
        users = list(db.users.find({}, {"_id": 0}))
        return jsonify(users)

    @bp.post("")
    def create_user():
        db = get_db()
        data = request.get_json(force=True)

        # very light validation
        username = (data.get("username") or "").strip()
        if not username:
            return jsonify({"error": "username is required"}), 400

        db.users.insert_one({"username": username})
        return jsonify({"created": True}), 201

    return bp
