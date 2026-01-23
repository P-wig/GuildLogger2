from flask import Blueprint, jsonify

def get_blueprint():
    bp = Blueprint("root", __name__)

    @bp.get("/")
    def api_info():
        return jsonify({
            "message": "Cloud Native Team Project API",
            "version": "1.0",
            "endpoints": {
                "health": "/api/health",
                "users": "/api/users"
            }
        })

    return bp