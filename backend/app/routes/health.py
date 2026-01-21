from flask import Blueprint, jsonify

def get_blueprint():
    bp = Blueprint("health", __name__, url_prefix="/api")

    @bp.get("/health")
    def health():
        return jsonify({"ok": True})

    return bp
