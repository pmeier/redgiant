import contextlib
import time
from typing import Any

import httpx
from anyio import Lock
from httpx_ws import AsyncWebSocketSession, aconnect_ws

from . import schema
from ._compat import typing_Self as Self
from ._i18n import Translator


class RedGiant:
    def __init__(self, host: str, *, language: schema.Language = "en_US"):
        self._host = host
        self._language = language
        self._translator = Translator(host, language=language)

        self._exit_stack: contextlib.AsyncExitStack()
        self._ws: AsyncWebSocketSession
        self._lock = Lock()
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
                    )
                ),
                client,
            )
        )

        self._token = (await self._send("connect")).data["token"]

    async def close(self) -> None:
        await self._exit_stack.aclose()

    async def get_devices(self) -> list[schema.Device]:
        response = await self._send("devicelist", is_check_token="0", type="0")
        return [
            schema.Device.model_validate(device) for device in response.data["list"]
        ]

    async def get_data(
        self, *, device_id: int, timestamp: int | None = None
    ) -> list[schema.Datapoint]:
        if timestamp is None:
            timestamp = int(time.time())
        response = await self._send("real", dev_id=str(device_id), time123456=timestamp)
        context = {"translator": self._translator}
        return [
            schema.Datapoint.model_validate(datapoint, context=context)
            for datapoint in response.data["list"]
        ]

    async def _send(self, service: str, **kwargs: Any) -> schema.Response:
        async with self._lock:
            await self._ws.send_json(
                {
                    "lang": self._language,
                    "token": self._token,
                    "service": service,
                    **kwargs,
                }
            )
            data = await self._ws.receive_json()
        response = schema.Response.model_validate(data)
        assert response.code == 1
        return response

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

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
