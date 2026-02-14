from __future__ import annotations

from flask import Blueprint, jsonify, request
from pydantic import ValidationError
from pymongo.collection import Collection

from app.db import get_db
from app.mongo_utils import serialize_doc, to_object_id
from app.schemas.hardware import (
    HardwareCheckin,
    HardwareCheckout,
    HardwareCreate,
    HardwareUpdate,
)

bp = Blueprint("hardware", __name__)


def hardware_col() -> Collection:
    return get_db()["hardware"]


def _validation_error_response(exc: ValidationError):
    """Convert a Pydantic ValidationError into a 400 JSON response."""
    errors = []
    for e in exc.errors():
        field = ".".join(str(loc) for loc in e["loc"])
        errors.append({"field": field, "message": e["msg"]})
    return jsonify({"error": "Validation failed", "details": errors}), 400


# ── List hardware ────────────────────────────────────────────────
@bp.get("")
def list_hardware():
    """Return all hardware sets, optionally filtered by assignedProject."""
    query: dict = {}
    assigned_project = request.args.get("assignedProject")

    if assigned_project:
        query["assignedProjects"] = assigned_project

    docs = [serialize_doc(d) for d in hardware_col().find(query).limit(200)]
    return jsonify(docs)


# ── Create hardware ──────────────────────────────────────────────
@bp.post("")
def create_hardware():
    data = request.get_json(silent=True) or {}

    try:
        body = HardwareCreate(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    # Check uniqueness on hardwareName
    if hardware_col().find_one({"hardwareName": body.hardwareName}):
        return jsonify({"error": "hardwareName already exists"}), 409

    doc = {
        "hardwareName": body.hardwareName,
        "capacity": body.capacity,
        "available": body.capacity,  # starts fully stocked
        "assignedProjects": [],
    }

    res = hardware_col().insert_one(doc)
    saved = hardware_col().find_one({"_id": res.inserted_id})
    if not saved:
        return jsonify({"error": "Failed to create hardware"}), 500

    return jsonify(serialize_doc(saved)), 201


# ── Get single hardware ─────────────────────────────────────────
@bp.get("/<hardware_id>")
def get_hardware(hardware_id: str):
    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    doc = hardware_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Not found"}), 404
    return jsonify(serialize_doc(doc))


# ── Partial update ───────────────────────────────────────────────
@bp.patch("/<hardware_id>")
def update_hardware(hardware_id: str):
    data = request.get_json(silent=True) or {}

    try:
        body = HardwareUpdate(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    updates = body.model_dump(exclude_none=True)
    if not updates:
        return jsonify({"error": "No valid fields to update"}), 400

    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = hardware_col().update_one({"_id": _id}, {"$set": updates})
    if res.matched_count == 0:
        return jsonify({"error": "Not found"}), 404

    doc = hardware_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Failed to retrieve hardware"}), 500
    return jsonify(serialize_doc(doc))


# ── Delete hardware ──────────────────────────────────────────────
@bp.delete("/<hardware_id>")
def delete_hardware(hardware_id: str):
    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    res = hardware_col().delete_one({"_id": _id})
    if res.deleted_count == 0:
        return jsonify({"error": "Not found"}), 404
    return "", 204


# ── Checkout hardware (stub) ────────────────────────────────────
@bp.post("/<hardware_id>/checkout")
def checkout_hardware(hardware_id: str):
    """Stub – validates the request body but returns 501."""
    data = request.get_json(silent=True) or {}

    try:
        HardwareCheckout(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    return jsonify({"error": "Not implemented"}), 501


# ── Checkin hardware (stub) ─────────────────────────────────────
@bp.post("/<hardware_id>/checkin")
def checkin_hardware(hardware_id: str):
    """Stub – validates the request body but returns 501."""
    data = request.get_json(silent=True) or {}

    try:
        HardwareCheckin(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    return jsonify({"error": "Not implemented"}), 501
