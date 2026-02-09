from __future__ import annotations

from flask import Blueprint, jsonify, request
from pymongo.collection import Collection

from app.db import get_db
from app.mongo_utils import serialize_doc, to_object_id

bp = Blueprint("projects", __name__)


def projects_col() -> Collection:
    return get_db()["projects"]


@bp.get("")
def list_projects():
    owner_user_id = request.args.get("ownerUserId") # read ownerUserId Filter
    query = {} # default query
    if owner_user_id:
        query = {"ownerUserId": owner_user_id} # if ownerUserId is provided, filter by it
    docs = [serialize_doc(d) for d in projects_col().find(query).limit(200)]
    return jsonify(docs) # return json list


@bp.post("")
def create_project():
    data = request.get_json(silent=True) or {}
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400

    project_id = data.get("projectId")
    name = data.get("name")
    description = data.get("description")
    owner_user_id = data.get("ownerUserId")

    missing_fields = []
    if not project_id or not str(project_id).strip(): #validate projectId
        missing_fields.append("projectId")
    if not name or not str(name).strip(): #validate name
        missing_fields.append("name")
    if not description or not str(description).strip(): #validate description
        missing_fields.append("description")

    if missing_fields:
        return jsonify({"error": f"Missing or empty fields: {', '.join(missing_fields)}"}), 400
    if projects_col().find_one({"projectId": project_id}):
        return jsonify({"error": "projectId exists"}), 409

    # build project document
    doc = {
           "projectId": str(project_id).strip(),
           "name": str(name).strip(),
           "description": str(description).strip(),
        }
    if owner_user_id and str(owner_user_id).strip():
        doc["ownerUserId"] = str(owner_user_id).strip()
    res = projects_col().insert_one(doc)
    saved_doc = projects_col().find_one({"_id":res.inserted_id})
    if not saved_doc:
        return jsonify({"error": "Failed to create project"}), 500
    return jsonify(serialize_doc(saved_doc)), 201



@bp.get("/<project_id>")
def get_project(project_id: str):
    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Not found"}), 404
    return jsonify(serialize_doc(doc))


@bp.put("/<project_id>")
def replace_project(project_id: str):
    data = request.get_json(silent=True) or {}
    if not isinstance(data, dict):
        return jsonify({"error": "Expected JSON object body"}), 400
    data.pop("_id", None)

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().replace_one({"_id": _id}, data, upsert=False)
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


@bp.patch("/<project_id>")
def update_project(project_id: str):
    data = request.get_json(silent=True) or {}
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400
    data.pop("_id", None)

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().update_one({"_id": _id}, {"$set": data})
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


@bp.delete("/<project_id>")
def delete_project(project_id: str):
    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().delete_one({"_id": _id})
    if res.deleted_count == 0:
        return jsonify({"error": "Not found"}), 404
    return "", 204
