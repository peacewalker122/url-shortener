import sqlite3


class Database:
    def __init__(self, filename):
        self.filename = filename
        self.pool: sqlite3.Connection
        self._connect()
        self._migrate()

    def _connect(self):
        self.pool = sqlite3.connect(
            self.filename, check_same_thread=False, factory=sqlite3.Connection
        )
        self.pool.row_factory = sqlite3.Row

    def _migrate(self):
        cursor = self.pool.cursor()
        cursor.execute("""
        CREATE TABLE IF NOT EXISTS urls (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            long_url TEXT NOT NULL,
            short_url TEXT UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
        """)
        cursor.close()
        self.pool.commit()

    def get_conn(self):
        if self.pool is None:
            self._connect()
        return self.pool.cursor()

    def close(self):
        if self.pool:
            self.pool.close()
