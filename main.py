import time
from flask import jsonify, request, g
import structlog
from config import app, base62_encode, generate_big_int
from err import InvalidAPIUsage
from db import Database
from contextlib import closing

db = Database(filename="test.db")
log = structlog.get_logger()


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
    g.start_time = time.time()
    log.debug("request", method=request.method, path=request.path)


@app.after_request
def after_request(response):
    latency = (time.time() - g.start_time) * 1000
    log.debug("response", status_code=str(response.status_code), latency=f"{latency}ms")
    return response


if __name__ == "__main__":
    app.run()
