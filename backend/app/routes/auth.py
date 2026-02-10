"""
Docstring for backend.app.routes.auth
"""
from flask import Blueprint, jsonify, request

from app.db import get_db

bp = Blueprint("auth", __name__)
def _encrypt(input_text: str, num_shift: int, dir_shift: int) -> str:
    """
    Encrypt using a simple cyclic cipher algorithim.
    
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

    # Algorithim
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
    userid = data.get("userid")
    password = data.get("password")
    # Validate no missing fields
    if userid is None or password is None:
        return jsonify({"ok": False, "error": "Missing userid or password"}), 400

    encrypt_userid = _encrypt(userid, 3, 1)
    encrypt_password = _encrypt(password, 3, 1)

    db = get_db()
    user = db["users"].find_one({
        "userid": encrypt_userid,
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
            "userid": userid  # Return unencrypted for client use
        }
    }), 200
