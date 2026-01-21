from flask import Flask
from flask_cors import CORS
from dotenv import load_dotenv
from app import db
from app.config import Config
from app.routes.health import bp as health_bp
from app.routes.users import bp as users_bp


def create_app():
    load_dotenv()

    app = Flask(__name__)
    app.config.from_object(Config)

    # Allow local dev React to call Flask
    CORS(app, resources={r"/api/*": {"origins": ["http://localhost:5173"]}})

    app.register_blueprint(health_bp)
    app.register_blueprint(users_bp)

    db.init_app(app)
    return app
