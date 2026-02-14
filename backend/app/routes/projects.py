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


def _validation_error_response(exc: ValidationError):
    """Convert a Pydantic ValidationError into a 400 JSON response."""
    errors = []
    for e in exc.errors():
        field = ".".join(str(loc) for loc in e["loc"])
        errors.append({"field": field, "message": e["msg"]})
    return jsonify({"error": "Validation failed", "details": errors}), 400


# ── List projects ────────────────────────────────────────────────
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


# ── Create project ───────────────────────────────────────────────
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


# ── Get single project ──────────────────────────────────────────
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


# ── Replace project ─────────────────────────────────────────────
@bp.put("/<project_id>")
def replace_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectCreate(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    replacement = {
        "projectId": body.projectId,
        "projectName": body.projectName,
        "description": body.description,
        "ownerUserId": body.ownerUserId,
    }

    res = projects_col().replace_one({"_id": _id}, replacement, upsert=False)
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


# ── Partial update ───────────────────────────────────────────────
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

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().update_one({"_id": _id}, {"$set": updates})
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


# ── Join project ─────────────────────────────────────────────────
@bp.post("/<project_id>/join")
def join_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectJoin(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().update_one(
        {"_id": _id},
        {"$addToSet": {"assignedUsers": body.userId}},
    )
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = projects_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve project"}), 500
    return jsonify(serialize_doc(doc))


# ── Leave project ────────────────────────────────────────────────
@bp.post("/<project_id>/leave")
def leave_project(project_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = ProjectLeave(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        _id = to_object_id(project_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = projects_col().update_one(
        {"_id": _id},
        {"$pull": {"assignedUsers": body.userId}},
    )
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    return jsonify({"ok": True}), 200


# ── Delete project ───────────────────────────────────────────────
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
