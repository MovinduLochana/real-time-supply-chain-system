"""OpenTelemetry tracing setup."""
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import SimpleSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import SimpleMetricReader
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import Resource
from typing import Optional
from config import settings
import logging


logger = logging.getLogger(__name__)


def setup_tracing() -> Optional[trace.Tracer]:
    """Setup OpenTelemetry tracing with Jaeger exporter."""
    if not settings.OTEL_ENABLED:
        logger.info("OpenTelemetry tracing is disabled")
        return None
    
    try:
        resource = Resource.create({
            "service.name": settings.SERVICE_NAME,
            "service.version": "2.0.0"
        })
        
        jaeger_exporter = JaegerExporter(
            agent_host_name=settings.JAEGER_HOST,
            agent_port=settings.JAEGER_PORT,
        )
        
        trace_provider = TracerProvider(resource=resource)
        trace_provider.add_span_processor(SimpleSpanProcessor(jaeger_exporter))
        trace.set_tracer_provider(trace_provider)
        
        logger.info(f"OpenTelemetry tracing enabled: {settings.JAEGER_HOST}:{settings.JAEGER_PORT}")
        return trace.get_tracer(__name__)
    except Exception as e:
        logger.error(f"Failed to setup tracing: {e}")
        return None


def setup_metrics() -> Optional[metrics.Meter]:
    """Setup OpenTelemetry metrics."""
    if not settings.OTEL_ENABLED:
        return None
    
    try:
        resource = Resource.create({
            "service.name": settings.SERVICE_NAME,
        })
        
        metric_reader = SimpleMetricReader()
        meter_provider = MeterProvider(resource=resource, metric_readers=[metric_reader])
        metrics.set_meter_provider(meter_provider)
        
        logger.info("OpenTelemetry metrics enabled")
        return metrics.get_meter(__name__)
    except Exception as e:
        logger.error(f"Failed to setup metrics: {e}")
        return None
