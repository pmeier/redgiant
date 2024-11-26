import contextlib
import time
import uuid
from typing import Any, cast

import anyio
import httpx
from httpx_ws import AsyncWebSocketSession, aconnect_ws

from . import schema
from ._compat import typing_Self as Self


class Translator:
    def __init__(self, host: str, *, language: schema.Language) -> None:
        response = httpx.get(
            httpx.URL(scheme="http", host=host, path=f"/i18n/{language}.properties")
        ).raise_for_status()
        self._translations = dict(line.split("=", 1) for line in response.iter_lines())

    def __call__(self, code: str) -> str:
        return self._translations.get(code, code)


class RedGiant:
    def __init__(self, host: str, *, language: schema.Language = "en_US"):
        self._host = host
        self._language = language
        self._translator = Translator(host, language=language)

        self._should_close = False
        self._exit_stack: contextlib.AsyncExitStack
        self._ws: AsyncWebSocketSession
        self._last_response = 0.0
        self._lock = anyio.Lock()
        self._token = ""

    async def connect(self) -> None:
        self._exit_stack = contextlib.AsyncExitStack()

        client = await self._exit_stack.enter_async_context(
            httpx.AsyncClient(event_hooks={"request": [self._case_headers]})
        )

        self._ws = await self._exit_stack.enter_async_context(
            aconnect_ws(
                str(
                    httpx.URL(
                        scheme="ws",
                        host=self._host,
                        port=8082,
                        path="/ws/home/overview",
                    ),
                ),
                client,
                keepalive_ping_interval_seconds=None,
            )
        )

        background_tasks = await self._exit_stack.enter_async_context(
            anyio.create_task_group()
        )
        background_tasks.start_soon(self._keepalive_ping)

        self._token = (await self._send("connect")).data["token"]

    async def close(self) -> None:
        self._should_close = True
        await self._exit_stack.aclose()

    async def _keepalive_ping(self) -> None:
        while not self._should_close:
            if time.time() - self._last_response >= 10.0:
                await self.ping()
            else:
                await anyio.sleep(1.0)

    async def ping(self) -> None:
        await self._send_json(
            {"lang": "zh_cn", "service": "ping", "token": ":", "id": str(uuid.uuid4())}
        )

    async def get_devices(self) -> list[schema.Device]:
        response = await self._send("devicelist", is_check_token="0", type="0")
        return [
            schema.Device.model_validate(device) for device in response.data["list"]
        ]

    async def get_data(self, *, device_id: int) -> list[schema.Datapoint]:
        response = await self._send(
            "real", dev_id=str(device_id), time123456=int(time.time())
        )
        datapoints = []
        for data in response.data["list"]:
            raw = schema.RawDatapoint.model_validate(data)

            description = self._translator(raw.i18n_code)

            value: float | str
            try:
                value = float(raw.value)
            except ValueError:
                value = self._translator(raw.value)

            datapoints.append(
                schema.Datapoint(
                    i18n_code=raw.i18n_code,
                    description=description,
                    value=value,
                    unit=raw.unit,
                )
            )
        return datapoints

    async def _send(self, service: str, **kwargs: Any) -> schema.Response:
        return schema.Response.model_validate(
            await self._send_json(
                {
                    "lang": self._language,
                    "token": self._token,
                    "service": service,
                    **kwargs,
                }
            )
        )

    async def _send_json(self, data: dict[str, Any]) -> dict[str, Any]:
        async with self._lock:
            await self._ws.send_json(data)
            event = cast(dict[str, Any], await self._ws.receive_json())
        self._last_response = time.time()
        assert event["result_code"] == 1
        assert event["result_msg"] == "success"
        return event

    # HTTP headers should be treated case-insensitive, but the Sungrow server errors
    # unless the headers below are not send in the exact casing
    _CASE_SENSITIVE_HEADERS = ["Sec-WebSocket-Key", "Connection", "Upgrade"]

    @classmethod
    async def _case_headers(cls, request: httpx.Request) -> None:
        # This looks like it does nothing, but request.headers is *not* a plain
        # dictionary. Accessing the data, e.g. by calling .pop(), is case-insensitive.
        # Thus, we are removing the header in whatever casing the library has set it
        # and re-insert it in the casing we need.
        for header in cls._CASE_SENSITIVE_HEADERS:
            request.headers[header] = request.headers.pop(header)

    async def __aenter__(self) -> Self:
        await self.connect()
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):  # type: ignore[no-untyped-def]
        await self.close()
