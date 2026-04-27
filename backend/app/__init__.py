from __future__ import annotations
import atexit

from flask import Flask
from flask_cors import CORS

from .config import Config
from .db import close_mongo, init_mongo, get_db
from .routes.auth import bp as auth_bp
from .routes.health import bp as health_bp
from .routes.projects import bp as projects_bp
from .routes.root import bp as root_bp
from .routes.users import bp as users_bp


def create_app() -> Flask:
    app = Flask(__name__)
    app.config.from_object(Config)

    CORS(
        app,
        resources={r"/api/*": {"origins": app.config["CORS_ORIGINS"]}},
    )

    init_mongo(app)

    atexit.register(lambda: close_mongo(app))

    app.register_blueprint(root_bp)
    app.register_blueprint(health_bp)
    app.register_blueprint(auth_bp, url_prefix="/api/auth")
    app.register_blueprint(users_bp, url_prefix="/api/users")
    app.register_blueprint(projects_bp, url_prefix="/api/projects")

    return app
