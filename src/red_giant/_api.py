import contextlib
from typing import AsyncIterator

from fastapi import APIRouter, FastAPI, Response, status

from red_giant import RedGiant, __version__, schema

__all__ = ["make_app"]


def make_app(sungrow_host: str, language: schema.Language) -> FastAPI:
    rg = RedGiant(sungrow_host, language=language)

    @contextlib.asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        async with rg:
            yield

    app = FastAPI(name="red-giant", version=__version__, lifespan=lifespan)

    @app.get("/health")
    async def health() -> Response:
        return Response(status_code=status.HTTP_200_OK)

    app.include_router(make_api_router(rg), prefix="/api")

    return app


def make_api_router(rg: RedGiant) -> APIRouter:
    router = APIRouter(tags=["API"])

    @router.get("/devices")
    async def get_devices() -> list[schema.Device]:
        print(rg._ws.connection._state)
        return await rg.get_devices()

    return router
