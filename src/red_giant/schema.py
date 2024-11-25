import contextlib
from typing import Annotated, Any, Literal

from pydantic import AfterValidator, BaseModel, Field, ValidationInfo, field_validator

__all__ = ["Language", "TranslatedStr", "Response", "Device"]

Language = Literal["zh_CN", "en_US", "de_DE", "nl_NL", "pl_PL"]


def _translate(v: str, info: ValidationInfo) -> str:
    translator = info.context.get("translator")
    if translator is not None:
        v = translator(v)
    return v


TranslatedStr = Annotated[str, AfterValidator(_translate)]


class Response(BaseModel):
    code: int = Field(alias="result_code")
    message: str = Field(alias="result_msg")
    data: dict[str, Any] = Field(alias="result_data")


class Device(BaseModel):
    id: int = Field(alias="dev_id")
    code: int = Field(alias="dev_code")
    type: int = Field(alias="dev_type")
    protocol: int = Field(alias="dev_procotol")
    serial_number: str = Field(alias="dev_sn")
    name: str = Field(alias="dev_name")
    model: str = Field(alias="dev_model")
    special: str = Field(alias="dev_special")

    inv_type: int
    port_name: str
    physical_address: int = Field(alias="phys_addr")
    logical_address: int = Field(alias="logc_addr")
    link_status: int
    init_status: int


class Datapoint(BaseModel):
    name: TranslatedStr = Field(alias="data_name")
    value: Any = Field(alias="data_value")
    unit: str = Field(alias="data_unit")

    @field_validator("value")
    @classmethod
    def _try_cast_value(cls, value: Any) -> Any:
        if not isinstance(value, str):
            return value

        with contextlib.suppress(ValueError):
            return int(value)

        with contextlib.suppress(ValueError):
            return float(value)

        return value
