FROM python:3.10 as builder

# Install poetry
RUN pip install poetry
RUN mkdir -p /app

ENV POETRY_NO_INTERACTION=1 \
  POETRY_VIRTUALENVS_IN_PROJECT=1 \
  POETRY_VIRTUALENVS_CREATE=1 \
  POETRY_CACHE_DIR=/tmp/poetry_cache

COPY . /app
WORKDIR /app

RUN poetry install --no-root && rm -rf $POETRY_CACHE_DIR

FROM python:3.10-slim-buster as runtime

ENV VIRTUAL_ENV=/app/.venv \
  PATH="/app/.venv/bin:$PATH"
COPY --from=builder ${VIRTUAL_ENV} ${VIRTUAL_ENV}

# RUN pip install --no-cache-dir psycopg2-binary

COPY --from=builder /app /app
WORKDIR /app

EXPOSE 8000
# CMD ["python3", "manage.py", "runserver", "0.0.0.0:8000"]

CMD granian --interface wsgi main:app --host 0.0.0.0 --port 8000 --workers 4 --threading-mode workers --http auto --log --log-level critical
