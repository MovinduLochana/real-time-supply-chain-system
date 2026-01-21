"""ML Pipelines Service - FastAPI application for ML model serving."""

from contextlib import asynccontextmanager
from typing import AsyncGenerator

import structlog
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from prometheus_client import make_asgi_app

from .config import Settings
from .database import DatabasePool
from .routes import demand, health, routes
from .telemetry import setup_telemetry

logger = structlog.get_logger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """Application lifespan management."""
    settings = Settings()

    # Setup telemetry
    setup_telemetry(settings.telemetry)

    logger.info("Starting ML Pipelines Service", version="0.1.0")

    # Initialize database pool
    db_pool = DatabasePool(settings.database)
    await db_pool.initialize()
    app.state.db_pool = db_pool

    # Store settings in app state
    app.state.settings = settings

    yield

    # Cleanup
    await db_pool.close()
    logger.info("ML Pipelines Service shutdown complete")


def create_app() -> FastAPI:
    """Create and configure the FastAPI application."""
    app = FastAPI(
        title="ML Pipelines Service",
        description="Machine learning pipelines for logistics",
        version="0.1.0",
        lifespan=lifespan,
    )

    # CORS middleware
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # Include routers
    app.include_router(health.router, tags=["health"])
    app.include_router(demand.router, prefix="/api/v1/demand", tags=["demand"])
    app.include_router(routes.router, prefix="/api/v1/routes", tags=["routes"])

    # Mount Prometheus metrics
    metrics_app = make_asgi_app()
    app.mount("/metrics", metrics_app)

    # Instrument with OpenTelemetry
    FastAPIInstrumentor.instrument_app(app)

    return app


app = create_app()
