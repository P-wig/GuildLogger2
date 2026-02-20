from __future__ import annotations
import atexit

import atexit

from flask import Flask
from flask_cors import CORS

from .config import Config
from .db import close_mongo, init_mongo, seed_hardware, get_db
from .routes.auth import bp as auth_bp
from .routes.hardware import bp as hardware_bp
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

    with app.app_context():
        db = get_db()
        seed_hardware(db)
        app.logger.info("Hardware seeded successfully")

    atexit.register(lambda: close_mongo(app))

    app.register_blueprint(root_bp)
    app.register_blueprint(health_bp)
    app.register_blueprint(auth_bp, url_prefix="/api/auth")
    app.register_blueprint(users_bp, url_prefix="/api/users")
    app.register_blueprint(projects_bp, url_prefix="/api/projects")
    app.register_blueprint(hardware_bp, url_prefix="/api/hardware")

    return app
