import time
from fastapi import Depends, FastAPI, Request, HTTPException, Response
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session
from sqlalchemy.sql import text
from config import base62_encode, generate_big_int
from db import get_db
from err import InvalidAPIUsage
import structlog
import asyncio

app = FastAPI()
log = structlog.get_logger()

# CORS middleware for allowing cross-origin requests
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# @app.on_event("startup")
# async def startup(db: Session = Depends(get_db)):
#     db.execute(
#         text(
#             "CREATE TABLE IF NOT EXISTS urls (id INTEGER PRIMARY KEY, long_url TEXT, short_url TEXT)"
#         )
#     )


@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    start_time = time.time()
    log.debug("request", method=request.method, path=request.url.path)
    try:
        response = await call_next(request)
    except asyncio.CancelledError:
        log.warning("request cancelled", method=request.method, path=request.url.path)
        raise
    process_time = (time.time() - start_time) * 1000
    log.debug(
        "response", status_code=response.status_code, latency=f"{process_time:.2f}ms"
    )
    return response


@app.get("/shorten/{key}")
async def get_url(key: str, db: Session = Depends(get_db)):
    result = db.execute(
        text("SELECT long_url FROM urls WHERE short_url = :key"),
        {"key": key},
    )
    url = result.fetchone()
    if url is None:
        raise HTTPException(status_code=404, detail="URL not found")
    return {"url": url[0]}


@app.post("/shorten")
async def create_short_url(request: Request, db: Session = Depends(get_db)):
    json_req = await request.json()
    long_url = json_req.get("url")
    if long_url is None:
        raise HTTPException(status_code=400, detail="Missing url parameter")

    id = generate_big_int()
    shorturl = base62_encode(id)

    db.execute(
        text(
            "INSERT INTO urls (id, long_url, short_url) VALUES (:id, :long_url, :short_url)"
        ),
        {"id": id, "long_url": long_url, "short_url": shorturl},
    )
    db.commit()

    return {"url": shorturl}


@app.exception_handler(InvalidAPIUsage)
async def invalid_api_usage_handler(request: Request, exc: InvalidAPIUsage):
    return JSONResponse(
        status_code=exc.status_code,
        content=exc.to_dict(),
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
