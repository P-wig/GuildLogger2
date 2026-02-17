from __future__ import annotations

from flask import Blueprint, jsonify, request
from pydantic import ValidationError
from pymongo.collection import Collection

from app.db import get_db
from app.mongo_utils import serialize_doc, to_object_id
from app.schemas.projects import ProjectCreate, ProjectJoin, ProjectLeave, ProjectUpdate

bp = Blueprint("projects", __name__)


def projects_col() -> Collection:
    return get_db()["projects"]


def hardware_col() -> Collection:
    return get_db()["hardware"]


def _validation_error_response(exc: ValidationError):
    """Convert a Pydantic ValidationError into a 400 JSON response."""
    errors = []
    for e in exc.errors():
        field = ".".join(str(loc) for loc in e["loc"])
        errors.append({"field": field, "message": e["msg"]})
    return jsonify({"error": "Validation failed", "details": errors}), 400


@bp.get("")
def list_projects():
    """Return all projects, optionally filtered by ownerUserId or assignedUser."""
    query: dict = {}
    owner = request.args.get("ownerUserId")
    assigned = request.args.get("assignedUser")

    if owner:
        query["ownerUserId"] = owner
    if assigned:
        query["assignedUsers"] = assigned  # Mongo matches if value is in array

    docs = [serialize_doc(d) for d in projects_col().find(query).limit(200)]
    return jsonify(docs)


@bp.post("")
def create_project():
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectCreate(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    # Check uniqueness
    if projects_col().find_one({"projectId": body.projectId}):
        return jsonify({"error": "projectId already exists"}), 409

    doc = {
        "projectId": body.projectId,
        "projectName": body.projectName,
        "description": body.description,
        "ownerUserId": body.ownerUserId,
        "assignedUsers": [body.ownerUserId],
        "assignedHardware": [],
    }

    res = projects_col().insert_one(doc)
    saved = projects_col().find_one({"_id": res.inserted_id})
    if not saved:
        return jsonify({"error": "Failed to create project"}), 500

    return jsonify(serialize_doc(saved)), 201


@bp.get("/<project_id>")
def get_project(project_id: str):
    doc = projects_col().find_one({"projectId": project_id})
    if not doc:
        return jsonify({"error": "Not found"}), 404
    return jsonify(serialize_doc(doc))


@bp.patch("/<project_id>")
def update_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectUpdate(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    updates = body.model_dump(exclude_none=True)
    if not updates:
        return jsonify({"error": "No valid fields to update"}), 400

    res = projects_col().update_one({"projectId": project_id}, {"$set": updates})
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"projectId": project_id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


@bp.post("/<project_id>/join")
def join_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectJoin(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    res = projects_col().update_one(
        {"projectId": project_id},
        {"$addToSet": {"assignedUsers": body.userId}},
    )
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"projectId": project_id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


@bp.post("/<project_id>/leave")
def leave_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectLeave(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    res = projects_col().update_one(
        {"projectId": project_id},
        {"$pull": {"assignedUsers": body.userId}},
    )
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"projectId": project_id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


@bp.delete("/<project_id>")
def delete_project(project_id: str):
    # Owner verification
    user_id = request.args.get("userId")
    if not user_id:
        return jsonify({"error": "userId query parameter is required"}), 400

    doc = projects_col().find_one({"projectId": project_id})
    if not doc:
        return jsonify({"error": "Not found"}), 404

    if doc["ownerUserId"] != user_id:
        return jsonify({"error": "Only the project owner can delete this project"}), 403

    # release all checked-out hardware back to available
    for entry in doc.get("assignedHardware", []):
        hw_id = entry["hardwareId"]
        amount = entry["amount"]
        try:
            hardware_col().update_one(
                {"_id": to_object_id(hw_id)},
                {
                    "$inc": {"available": amount},
                    "$pull": {"assignedProjects": project_id},
                },
            )
        except Exception:
            pass

    projects_col().delete_one({"_id": doc["_id"]})
    return "", 204
