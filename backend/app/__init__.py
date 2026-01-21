from flask import Flask
from flask_cors import CORS
from dotenv import load_dotenv
from app import db
from app.config import Config
import os
import importlib
import inspect


def create_app():
    load_dotenv()

    app = Flask(__name__)
    app.config.from_object(Config)

    # Allow local dev React to call Flask
    CORS(app, resources={r"/api/*": {"origins": ["http://localhost:5173"]}})

    # find the routes directory
    routes_dir = os.path.join(os.path.dirname(__file__), 'routes')
    # begins to loop through all files in the routes directory
    for filename in os.listdir(routes_dir):
        # skip sub-directories and focus only on .py files (excluding __init__.py)
        filepath = os.path.join(routes_dir, filename)
        if os.path.isfile(filepath) and filename.endswith('.py') and filename != '__init__.py':
            # constructs the module name by removing the .py extension
            module_name = f"app.routes.{filename[:-3]}"
            # imports the module dynamically
            module = importlib.import_module(module_name)
            # checks module for get_blueprint function
            if hasattr(module, "get_blueprint"):
                app.register_blueprint(module.get_blueprint())
            else:
                print(f"Module {module_name} does not have a get_blueprint function")
        else:
            print(f"File {filename} does not match .py pattern or is __init__.py")

    db.init_app(app)
    return app
