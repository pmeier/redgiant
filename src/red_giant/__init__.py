try:
    from ._version import __version__
except ModuleNotFoundError:
    import warnings

    warnings.warn("red-giant was not properly installed!")
    del warnings

    __version__ = "UNKNOWN"

from . import schema
from ._red_giant import RedGiant
