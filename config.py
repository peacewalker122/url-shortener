import logging
from environs import Env
import random
import unittest
from flask import Flask

logging.basicConfig(
    format="%(asctime)s - %(levelname)s - %(message)s", level=logging.DEBUG
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
