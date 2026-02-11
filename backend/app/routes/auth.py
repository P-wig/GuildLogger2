"""
Docstring for backend.app.routes.auth
"""
from flask import Blueprint, jsonify, request

from app.db import get_db
from app.mongo_utils import serialize_doc
from app.routes.users import users_col

bp = Blueprint("auth", __name__)
def _encrypt(input_text: str, num_shift: int, dir_shift: int) -> str:
    """
    Encrypt using a simple cyclic cipher algorithm.
    
    :param input_text: text to encrypt
    :type input_text: str
    :param num_shift: number of shifts
    :type num_shift: int
    :param dir_shift: direction of shift
    :type dir_shift: int
    :return: (cryptedText) encrypted version of input_text
    :rtype: str
    """

    if not input_text.isascii():
        raise TypeError
    elif num_shift < 1:
        raise ValueError("NUM shift N must be >=1")
    elif dir_shift < -1 and dir_shift > 1:
        raise ValueError("Direction shift D must be either +1 or -1")

    forbidden_chars = (' ', '!') # ASCII 32, 33
    if any(char in forbidden_chars for char in input_text):
        raise ValueError("Input contains forbidden ASCII codes 32 or 33 (!/SPACE)")

    # Algorithm
    # Step 1: Reverse input text
    input_text = "".join(reversed(input_text))
    # Step 2: Shift all the ASCII characters
    # in reversed input text by num_shift * dir_shift positions
    shifted_chars = []
    for char in input_text:
        old_ascii = ord(char)
        new_ascii = old_ascii + (num_shift * dir_shift)
        # Ensure we stay within valid ASCII range (0-127)
        if new_ascii > 127:
            new_ascii = new_ascii - 128  # Wrap around
        elif new_ascii < 34:
            new_ascii = new_ascii + 128  # Wrap around
        shifted_chars.append(chr(new_ascii))  # Convert back to character

    return "".join(shifted_chars)

@bp.post("/login")
def login():
    """
    Docstring for login
    """
    data = request.get_json()
    userid = data.get("userId")
    password = data.get("password")
    # Validate no missing fields
    if userid is None or password is None:
        return jsonify({"ok": False, "error": "Missing userId or password"}), 400

    encrypt_userid = _encrypt(userid, 3, 1)
    encrypt_password = _encrypt(password, 3, 1)

    db = get_db()
    user = db["users"].find_one({
        "userId": encrypt_userid,
        "password": encrypt_password
    })
    # Check if user exists
    if user is None:
        return jsonify({"ok": False, "error": "Invalid credentials"}), 401
    # Successful login
    return jsonify({
        "ok": True,
        "message": "Login successful",
        "user": {
            "userId": userid  # Return unencrypted for client use
        }
    }), 200

@bp.post("/register")
def register():
    """
    Docstring for register
    """
    data = request.get_json(silent=True) or {}
    # minimal validation (extend later)
    if not isinstance(data, dict) or not data:
        return jsonify({"error": "Expected JSON object body"}), 400

    required_fields = ["userId", "password"]
    if not all(field in data for field in required_fields):
        missing_fields = [field for field in required_fields if field not in data]
        error_message = f"Missing fields: {', '.join(missing_fields)}"
        return jsonify({"error": error_message}), 400
    else:
        # Init return object
        user_data = {"userId": data["userId"]}
        # Create Hashed Password and UserID
        hash_userid = _encrypt(data["userId"], 3, 1)
        # Check if userId already exists
        if users_col().find_one({"userId": hash_userid}):
            return jsonify({"error": "userId already exists"}), 409
        else:
            hash_password = _encrypt(data["password"], 3, 1)
        data["userId"] = hash_userid
        data["password"] = hash_password
        res = users_col().insert_one(data)
        doc = users_col().find_one({"_id": res.inserted_id})

        if not doc:
            return jsonify({"error": "Failed to create user"}), 500
        else:
            return jsonify({"user":user_data}), 201