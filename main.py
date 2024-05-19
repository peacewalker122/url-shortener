from flask import jsonify, request
from config import app, base62_encode, generate_big_int
from err import InvalidAPIUsage
from db import Database
from contextlib import closing
import logging

db = Database(filename="test.db")


@app.route("/")
def hello():
    return "<p>Hello, World!</p>", 200


@app.route("/shorten/<key>", methods=["GET"])
def get(key):
    with closing(db.get_conn()) as cur:
        cur.execute("SELECT * FROM urls WHERE short_url = ?", (key,))
        row = cur.fetchone()
        if row is None:
            raise InvalidAPIUsage("Key does not exist", status_code=400)
        return {"url": row[1]}, 200


@app.route("/shorten", methods=["POST"])
def put():
    json_req = request.get_json()
    long_url = json_req.get("url")
    if long_url is None:
        raise InvalidAPIUsage("Missing url parameter", status_code=400)

    id = generate_big_int()
    shorturl = base62_encode(id)

    with closing(db.get_conn()) as cur:
        cur.execute(
            "INSERT INTO urls (id, long_url, short_url) VALUES (?, ?, ?)",
            (id, long_url, shorturl),
        )
        return {"url": shorturl}, 200


@app.errorhandler(InvalidAPIUsage)
def invalid_api_usage(e):
    return jsonify(e.to_dict()), e.status_code


@app.before_request
def before_request():
    logging.debug(f"Request: {request.method} {request.path}")


@app.after_request
def after_request(response):
    logging.debug(f"Response: {response.status_code} {response.headers}")
    return response


if __name__ == "__main__":
    app.run()
