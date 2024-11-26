from typing import Annotated, Optional

import httpx
import rich
import typer
import uvicorn

import red_giant

from ._api import make_app as make_api_app

app = typer.Typer(
    name="red-giant",
    invoke_without_command=True,
    no_args_is_help=True,
    add_completion=False,
    pretty_exceptions_enable=False,
)


def version_callback(value: bool) -> None:
    if value:
        rich.print(f"red-giant {red_giant.__version__} from {red_giant.__path__[0]}")
        raise typer.Exit()


@app.callback()
def _main(
    version: Annotated[
        Optional[bool],
        typer.Option(
            "--version", callback=version_callback, help="Show version and exit."
        ),
    ] = None,
) -> None:
    pass


DEFAULT_HOST = "127.0.0.1"
HostOption = Annotated[
    str,
    typer.Option(help="Host the REST API will be bound to", envvar="RED_GIANT_HOST"),
]

DEFAULT_PORT = 8000
PortOption = Annotated[
    int,
    typer.Option(help="Port the REST API will be bound to", envvar="RED_GIANT_PORT"),
]


@app.command(help="Check whether the REST API is healthy")
def healthcheck(
    *,
    host: HostOption = DEFAULT_HOST,
    port: PortOption = DEFAULT_PORT,
) -> None:
    def is_healthy() -> bool:
        try:
            return httpx.get(
                httpx.URL(scheme="http", host=host, port=port, path="/health")
            ).is_success
        except httpx.HTTPError:
            return False

    raise SystemExit(int(not is_healthy()))


@app.command(help="Serve the REST API")
def serve(
    *,
    sungrow_host: Annotated[
        str,
        typer.Option(
            show_default=False,
            help="Host of the Sungrow inverter",
            envvar="RED_GIANT_SUNGROW_HOST",
        ),
    ],
    language: Annotated[
        str, typer.Option(help="Language of the returned content.")
    ] = "en_US",
    host: HostOption = DEFAULT_HOST,
    port: PortOption = DEFAULT_PORT,
) -> None:
    uvicorn.run(
        make_api_app(
            sungrow_host,
            language,  # type: ignore[arg-type]
        ),
        host=host,
        port=port,
    )
