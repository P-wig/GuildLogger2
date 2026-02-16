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


def projects_col() -> Collection:
    return get_db()["projects"]


def _validation_error_response(exc: ValidationError):
    """Convert a Pydantic ValidationError into a 400 JSON response."""
    errors = []
    for e in exc.errors():
        field = ".".join(str(loc) for loc in e["loc"])
        errors.append({"field": field, "message": e["msg"]})
    return jsonify({"error": "Validation failed", "details": errors}), 400


@bp.get("")
def list_hardware():
    """Return all hardware sets, optionally filtered by assignedProject."""
    query: dict = {}
    assigned_project = request.args.get("assignedProject")

    if assigned_project:
        query["assignedProjects"] = assigned_project

    docs = [serialize_doc(d) for d in hardware_col().find(query).limit(200)]
    return jsonify(docs)


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


@bp.delete("/<hardware_id>")
def delete_hardware(hardware_id: str):
    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    doc = hardware_col().find_one({"_id": _id})
    if not doc:
        return jsonify({"error": "Not found"}), 404

    hw_id_str = str(_id)

    projects_col().update_many(
        {"assignedHardware.hardwareId": hw_id_str},
        {"$pull": {"assignedHardware": {"hardwareId": hw_id_str}}},
    )

    hardware_col().delete_one({"_id": _id})
    return "", 204


@bp.post("/<hardware_id>/checkout")
def checkout_hardware(hardware_id: str):
    """Check out units of hardware for a project."""
    data = request.get_json(silent=True) or {}

    try:
        body = HardwareCheckout(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    hw = hardware_col().find_one({"_id": _id})
    if not hw:
        return jsonify({"error": "Hardware not found"}), 404

    project = projects_col().find_one({"projectId": body.projectId})
    if not project:
        return jsonify({"error": "Project not found"}), 404

    if body.userId not in project.get("assignedUsers", []):
        return jsonify({"error": "User is not assigned to this project"}), 403

    if hw["available"] < body.amount:
        return (
            jsonify(
                {
                    "error": f"Insufficient availability. Only {hw['available']} units available"
                }
            ),
            400,
        )

    hw_id_str = str(_id)
    amount = int(body.amount)

    # Decrease available on hardware, track project
    hardware_col().update_one(
        {"_id": _id},
        {
            "$inc": {"available": -amount},
            "$addToSet": {"assignedProjects": body.projectId},
        },
    )

    # Update project's assignedHardware: increment if exists, else push new
    existing_entry = next(
        (
            e
            for e in project.get("assignedHardware", [])
            if e["hardwareId"] == hw_id_str
        ),
        None,
    )
    if existing_entry:
        projects_col().update_one(
            {"projectId": body.projectId, "assignedHardware.hardwareId": hw_id_str},
            {"$inc": {"assignedHardware.$.amount": body.amount}},
        )
    else:
        projects_col().update_one(
            {"projectId": body.projectId},
            {
                "$push": {
                    "assignedHardware": {"hardwareId": hw_id_str, "amount": body.amount}
                }
            },
        )

    updated_hw = hardware_col().find_one({"_id": _id})
    if not updated_hw:
        return jsonify({"error": "Failed to retrieve hardware"}), 500
    return jsonify(serialize_doc(updated_hw))


@bp.post("/<hardware_id>/checkin")
def checkin_hardware(hardware_id: str):
    """Check in units of hardware from a project."""
    data = request.get_json(silent=True) or {}

    try:
        body = HardwareCheckin(**data)
    except ValidationError as exc:
        return _validation_error_response(exc)

    try:
        _id = to_object_id(hardware_id)
    except Exception:
        return jsonify({"error": "Invalid id"}), 400

    hw = hardware_col().find_one({"_id": _id})
    if not hw:
        return jsonify({"error": "Hardware not found"}), 404

    project = projects_col().find_one({"projectId": body.projectId})
    if not project:
        return jsonify({"error": "Project not found"}), 404

    if body.userId not in project.get("assignedUsers", []):
        return jsonify({"error": "User is not assigned to this project"}), 403

    hw_id_str = str(_id)
    amount = int(body.amount)

    entry = next(
        (
            e
            for e in project.get("assignedHardware", [])
            if e["hardwareId"] == hw_id_str
        ),
        None,
    )
    if not entry:
        return (
            jsonify({"error": "This hardware is not checked out for this project"}),
            400,
        )

    if amount > entry["amount"]:
        return (
            jsonify(
                {
                    "error": f"Cannot check in {amount} units. Only {entry['amount']} checked out"
                }
            ),
            400,
        )

    new_available = min(hw["capacity"], hw["available"] + amount)
    hardware_col().update_one({"_id": _id}, {"$set": {"available": new_available}})

    if amount == entry["amount"]:

        projects_col().update_one(
            {"projectId": body.projectId},
            {"$pull": {"assignedHardware": {"hardwareId": hw_id_str}}},
        )
        hardware_col().update_one(
            {"_id": _id},
            {"$pull": {"assignedProjects": body.projectId}},
        )
    else:
        # Partially checked in
        projects_col().update_one(
            {"projectId": body.projectId, "assignedHardware.hardwareId": hw_id_str},
            {"$inc": {"assignedHardware.$.amount": -amount}},
        )

    updated_hw = hardware_col().find_one({"_id": _id})
    if not updated_hw:
        return jsonify({"error": "Failed to retrieve hardware"}), 500
    return jsonify(serialize_doc(updated_hw))
