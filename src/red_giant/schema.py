from typing import Annotated, Any, Literal

from pydantic import AfterValidator, BaseModel, Field, ValidationInfo

__all__ = ["Language", "TranslatedStr", "Response", "Device"]

Language = Literal["zh_CN", "en_US", "de_DE", "nl_NL", "pl_PL"]


def _translate(v: str, info: ValidationInfo) -> str:
    return v
    if not info.context:
        return v
    translator = info.context.get("translator")
    if translator is not None:
        v = translator(v)
    return v


TranslatedStr = Annotated[str, AfterValidator(_translate)]


class Response(BaseModel):
    code: int = Field(validation_alias="result_code")
    message: str = Field(validation_alias="result_msg")
    data: dict[str, Any] = Field(validation_alias="result_data")


class Device(BaseModel):
    id: int = Field(validation_alias="dev_id")
    code: int = Field(validation_alias="dev_code")
    type: int = Field(validation_alias="dev_type")
    protocol: int = Field(validation_alias="dev_procotol")
    serial_number: str = Field(validation_alias="dev_sn")
    name: str = Field(validation_alias="dev_name")
    model: str = Field(validation_alias="dev_model")
    special: str = Field(validation_alias="dev_special")

    inv_type: int
    port_name: str
    physical_address: int = Field(validation_alias="phys_addr")
    logical_address: int = Field(validation_alias="logc_addr")
    link_status: int
    init_status: int


class RawDatapoint(BaseModel):
    i18n_code: str = Field(validation_alias="data_name")
    value: str = Field(validation_alias="data_value")
    unit: str = Field(validation_alias="data_unit")


class Datapoint(BaseModel):
    i18n_code: str = Field(exclude=True)
    description: str
    value: float | str
    unit: str
