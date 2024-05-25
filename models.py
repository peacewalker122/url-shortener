from sqlalchemy import Column, Integer, String, TIMESTAMP, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.sql import func

Base = declarative_base()


class URL(Base):
    __tablename__ = "urls"

    id = Column(Integer, primary_key=True, index=True)
    long_url = Column(String, nullable=False)
    short_url = Column(String, unique=True, nullable=False)
    created_at = Column(TIMESTAMP, server_default=func.now())
