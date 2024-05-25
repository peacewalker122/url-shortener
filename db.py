from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.sql import text


DATABASE_URL = "sqlite:///../test.db"

engine = create_engine(DATABASE_URL)

SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

with engine.connect() as conn:
    conn.execute(
        text(
            "CREATE TABLE IF NOT EXISTS urls (id INTEGER PRIMARY KEY, long_url TEXT, short_url TEXT)"
        )
    )


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
