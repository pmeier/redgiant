import contextlib
from typing import AsyncIterator, cast

from fastapi import APIRouter, FastAPI, Request, Response, status
from prometheus_client import CollectorRegistry, Gauge
from prometheus_client.exposition import _bake_output

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
    app.include_router(make_live_data_router(rg), prefix="/live-data")

    return app


def make_api_router(rg: RedGiant) -> APIRouter:
    router = APIRouter(tags=["API"])

    @router.get("/devices")
    async def get_devices() -> list[schema.Device]:
        return await rg.get_devices()

    return router


def make_live_data_router(rg: RedGiant) -> APIRouter:
    router = APIRouter(tags=["Live data"])

    registry = CollectorRegistry(auto_describe=False)

    # Unfortunately, we cannot just mount the prometheus_client.make_asgi_app, because
    # we do not sample continuously. Thus, we have to sample when the request comes in.
    # Furthermore, prometheus_client does not support async and thus we actually have to
    # wrap internals.
    @router.get("/metrics")
    async def metrics(request: Request) -> Response:
        await sample()
        _, headers, output = _bake_output(  # type: ignore[no-untyped-call]
            registry,
            accept_header=request.headers.get("Accept"),
            accept_encoding_header=request.headers.get("Accept-Encoding"),
            params=request.query_params,
            disable_compression=False,
        )
        return Response(output, status_code=status.HTTP_200_OK, headers=dict(headers))

    initialized = False
    devices: list[schema.Device] = []
    gauges: dict[tuple[int, str], tuple[Gauge, float]] = {}

    async def sample() -> None:
        if not initialized:
            await initialize()
            return

        for device in devices:
            for datapoint in await rg.get_data(device_id=device.id):
                key = (device.id, datapoint.i18n_code)
                value = gauges.get(key)
                if value is None:
                    continue
                gauge, factor = value

                gauge.set(cast(float, datapoint.value) * factor)

    # Wirkleistung  -> Total active power / inverter
    # Batterieladeleistung ->
    # Batterieentladeleistung

    # Battery level -> battery level / battery
    # purchased power -> purchased power / inverter

    FOO = {
        "I18N_CONFIG_KEY_4060": "grid_power",
        "I18N_COMMON_TOTAL_ACTIVE_POWER": "pv_power",
        "I18N_COMMON_TOTAL_BACKUP_POWER_WLECIVPM": "battery_power",
        "I18N_COMMON_BATTERY_VOLTAGE": "battery_voltage",
        "I18N_COMMON_BATTERY_CURRENT": "battery_current",
        "I18N_COMMON_BATTERY_TEMPERATURE": "battery_temperature",
        "I18N_COMMON_REMAIN_BATTERY_POWER": "battery_level",
        "I18N_COMMON_BATTARY_HEALTH": "battery_health",
    }

    # See https://prometheus.io/docs/practices/naming/#base-units
    base_unit_conversions: dict[str, tuple[str, float]] = {
        "V": ("volts", 1),
        "A": ("amperes", 1),
        # https://www.compart.com/en/unicode/U+2103
        "℃": ("celcius", 1),
        "%": ("ratio", 1e-2),
        "kWh": ("joules", 3.6e6),
        "kW": ("watts", 1e3),
    }

    async def initialize():
        nonlocal devices
        devices[:] = await rg.get_devices()

        nonlocal gauges
        for device in devices:
            for datapoint in await rg.get_data(device_id=device.id):
                name = FOO.get(datapoint.i18n_code)
                if name is None:
                    continue

                base_unit, factor = base_unit_conversions.get(
                    datapoint.unit, (datapoint.unit, 1)
                )
                gauge = Gauge(
                    name=name, documentation=datapoint.description, unit=base_unit
                )
                gauge.set(cast(float, datapoint.value) * factor)

                registry.register(gauge)

                gauges[(device.id, datapoint.i18n_code)] = (gauge, factor)

        nonlocal initialized
        initialized = True

    return router
