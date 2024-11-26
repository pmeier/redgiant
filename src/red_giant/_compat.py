import sys

__all__ = ["typing_Self"]


def _typing_Self():  # type: ignore[no-untyped-def]
    if sys.version_info[:2] >= (3, 11):
        from typing import Self
    else:
        from typing_extensions import Self

    return Self


typing_Self = _typing_Self()  # type: ignore[no-untyped-call]
