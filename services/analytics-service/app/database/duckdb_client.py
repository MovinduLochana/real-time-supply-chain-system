"""DuckDB client and connection management."""
import duckdb
import logging
import os
from typing import Optional, Dict, List, Any
from pathlib import Path
from config import settings


logger = logging.getLogger(__name__)


class DuckDBClient:
    """DuckDB client for analytics."""
    
    def __init__(self, db_path: str = settings.DB_PATH):
        """Initialize DuckDB client."""
        self.db_path = db_path
        self.connection: Optional[duckdb.DuckDBPyConnection] = None
        self._ensure_directories()
    
    def _ensure_directories(self) -> None:
        """Ensure required directories exist."""
        db_dir = os.path.dirname(self.db_path)
        if db_dir and not os.path.exists(db_dir):
            os.makedirs(db_dir, exist_ok=True)
        
        parquet_dir = settings.DB_PARQUET_PATH
        if not os.path.exists(parquet_dir):
            os.makedirs(parquet_dir, exist_ok=True)
    
    def connect(self) -> duckdb.DuckDBPyConnection:
        """Create or return existing connection."""
        if self.connection is None:
            logger.info(f"Connecting to DuckDB: {self.db_path}")
            self.connection = duckdb.connect(self.db_path)
            self.connection.execute("INSTALL httpfs;")
            self.connection.execute("LOAD httpfs;")
        return self.connection
    
    def disconnect(self) -> None:
        """Close connection."""
        if self.connection is not None:
            self.connection.close()
            self.connection = None
            logger.info("Disconnected from DuckDB")
    
    def execute(self, query: str, params: Optional[List] = None) -> Any:
        """Execute a query."""
        try:
            conn = self.connect()
            if params:
                return conn.execute(query, params)
            return conn.execute(query)
        except Exception as e:
            logger.error(f"Query execution failed: {e}, Query: {query}")
            raise
    
    def fetch_all(self, query: str, params: Optional[List] = None) -> List[Dict[str, Any]]:
        """Fetch all results as list of dicts."""
        result = self.execute(query, params)
        return [dict(row) for row in result.fetchall()]
    
    def fetch_one(self, query: str, params: Optional[List] = None) -> Optional[Dict[str, Any]]:
        """Fetch single result."""
        result = self.execute(query, params)
        row = result.fetchone()
        return dict(row) if row else None
    
    def fetch_df(self, query: str, params: Optional[List] = None):
        """Fetch results as pandas DataFrame."""
        result = self.execute(query, params)
        return result.df()
    
    def create_table(self, table_name: str, schema: str) -> None:
        """Create a table."""
        try:
            self.execute(f"CREATE TABLE IF NOT EXISTS {table_name} ({schema})")
            logger.info(f"Table created: {table_name}")
        except Exception as e:
            logger.error(f"Failed to create table {table_name}: {e}")
            raise
    
    def insert_data(self, table_name: str, data: List[Dict[str, Any]]) -> None:
        """Insert data into table."""
        if not data:
            return
        
        try:
            import pandas as pd
            df = pd.DataFrame(data)
            conn = self.connect()
            conn.register(f"temp_{table_name}", df)
            conn.execute(f"INSERT INTO {table_name} SELECT * FROM temp_{table_name}")
            conn.unregister(f"temp_{table_name}")
            logger.debug(f"Inserted {len(data)} rows into {table_name}")
        except Exception as e:
            logger.error(f"Failed to insert data into {table_name}: {e}")
            raise
    
    def export_to_parquet(self, table_name: str, output_path: Optional[str] = None) -> str:
        """Export table to parquet."""
        if output_path is None:
            output_path = f"{settings.DB_PARQUET_PATH}/{table_name}_{pd.Timestamp.now().strftime('%Y%m%d_%H%M%S')}.parquet"
        
        try:
            self.execute(f"COPY {table_name} TO '{output_path}'")
            logger.info(f"Exported {table_name} to {output_path}")
            return output_path
        except Exception as e:
            logger.error(f"Failed to export {table_name}: {e}")
            raise
    
    def table_exists(self, table_name: str) -> bool:
        """Check if table exists."""
        try:
            result = self.fetch_one(
                f"SELECT name FROM information_schema.tables WHERE table_name = ?",
                [table_name]
            )
            return result is not None
        except Exception as e:
            logger.error(f"Failed to check table existence: {e}")
            return False
    
    def vacuum(self) -> None:
        """Optimize database."""
        try:
            self.execute("PRAGMA optimize")
            logger.info("Database optimized")
        except Exception as e:
            logger.warning(f"Failed to optimize database: {e}")


# Global client instance
_db_client: Optional[DuckDBClient] = None


def get_db_client() -> DuckDBClient:
    """Get or create global database client."""
    global _db_client
    if _db_client is None:
        _db_client = DuckDBClient()
    return _db_client


async def init_db() -> None:
    """Initialize database schema."""
    client = get_db_client()
    
    # Create tables if they don't exist
    if not client.table_exists("orders_fact"):
        from app.database.queries import CREATE_ORDERS_TABLE
        client.execute(CREATE_ORDERS_TABLE)
    
    if not client.table_exists("inventory_fact"):
        from app.database.queries import CREATE_INVENTORY_TABLE
        client.execute(CREATE_INVENTORY_TABLE)
    
    if not client.table_exists("customers_dim"):
        from app.database.queries import CREATE_CUSTOMERS_TABLE
        client.execute(CREATE_CUSTOMERS_TABLE)
    
    if not client.table_exists("forecasts"):
        from app.database.queries import CREATE_FORECASTS_TABLE
        client.execute(CREATE_FORECASTS_TABLE)
    
    logger.info("Database initialized successfully")


async def close_db() -> None:
    """Close database connection."""
    global _db_client
    if _db_client is not None:
        _db_client.disconnect()
        _db_client = None
