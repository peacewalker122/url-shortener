import logging
import structlog
import random
import unittest
from environs import Env
from flask import Flask

structlog.configure(
    processors=[
        structlog.contextvars.merge_contextvars,
        structlog.processors.add_log_level,
        structlog.processors.StackInfoRenderer(),
        structlog.dev.set_exc_info,
        structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S", utc=False),
        # structlog.dev.ConsoleRenderer(),
        structlog.processors.JSONRenderer(),
    ],
    wrapper_class=structlog.make_filtering_bound_logger(logging.NOTSET),
    context_class=dict,
    logger_factory=structlog.PrintLoggerFactory(),
    cache_logger_on_first_use=False,
)

env = Env()
env.read_env()
app = Flask(__name__)

app.config["ENV"] = env("FLASK_ENV")
app.config["DEBUG"] = env("DEBUG") == "1"
# app.config["SECRET_KEY"] = env("SECRET_KEY")


def base62_encode(val: int):
    chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

    ret = ""
    while val > 0:
        remainder = val % 62
        ret += chars[remainder]
        val //= 62

    return ret[::-1]


def generate_big_int():
    """
    Generate a random 64-bit integer.
    """
    return random.randrange(62**7, 62**8 - 1)


class TestEncode(unittest.TestCase):
    def test_encode(self):
        self.assertEqual(base62_encode(11157), "2TX")

    def test_generate_big_int(self):
        val = generate_big_int()
        print(base62_encode(val))


if __name__ == "__main__":
    unittest.main()
