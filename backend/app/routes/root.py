from flask import Blueprint, jsonify

bp = Blueprint("root", __name__)


@bp.get("/")
def root():
    return jsonify({"service": "backend", "status": "ok"})
