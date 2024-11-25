import httpx

from . import schema


class Translator:
    def __init__(self, host: str, *, language: schema.Language) -> None:
        response = httpx.get(
            httpx.URL(scheme="http", host=host, path=f"/i18n/{language}.properties")
        ).raise_for_status()
        self._i18n = dict(line.split("=", 1) for line in response.iter_lines())

    def __call__(self, s: str) -> str:
        return self._i18n.get(s, s)
