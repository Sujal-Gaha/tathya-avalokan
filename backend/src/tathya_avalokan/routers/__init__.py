from .health import router as health_router
from .instances import router as instances_router
from .projects import router as projects_router
from .query import router as query_router

__all__ = ["health_router", "instances_router", "projects_router", "query_router"]
