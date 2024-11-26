import httpx

from . import schema


class Translator:
    def __init__(self, host: str, *, language: schema.Language) -> None:
        response = httpx.get(
            httpx.URL(scheme="http", host=host, path=f"/i18n/{language}.properties")
        ).raise_for_status()
        self._translations = dict(line.split("=", 1) for line in response.iter_lines())
