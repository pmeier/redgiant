package errors

func (err *Error) Any(key string, value ) *Error {
    err.e.Any(key, value)
	return err
}

func (err *Error) Array(key string, value LogArrayMarshaler) *Error {
    err.e.Array(key, value)
	return err
}

func (err *Error) Bool(key string, value bool) *Error {
    err.e.Bool(key, value)
	return err
}

func (err *Error) Bools(key string, value ) *Error {
    err.e.Bools(key, value)
	return err
}

func (err *Error) Bytes(key string, value ) *Error {
    err.e.Bytes(key, value)
	return err
}

func (err *Error) Caller(value ) *Error {
    err.e.Caller(value)
	return err
}

func (err *Error) CallerSkipFrame(value int) *Error {
    err.e.CallerSkipFrame(value)
	return err
}

func (err *Error) Ctx(value Context) *Error {
    err.e.Ctx(value)
	return err
}

func (err *Error) Dict(key string, value ) *Error {
    err.e.Dict(key, value)
	return err
}

func (err *Error) Dur(key string, value Duration) *Error {
    err.e.Dur(key, value)
	return err
}

func (err *Error) Durs(key string, value ) *Error {
    err.e.Durs(key, value)
	return err
}

func (err *Error) EmbedObject(value LogObjectMarshaler) *Error {
    err.e.EmbedObject(value)
	return err
}

func (err *Error) Errs(key string, value ) *Error {
    err.e.Errs(key, value)
	return err
}

func (err *Error) Fields(value ) *Error {
    err.e.Fields(value)
	return err
}

func (err *Error) Float32(key string, value float32) *Error {
    err.e.Float32(key, value)
	return err
}

func (err *Error) Float64(key string, value float64) *Error {
    err.e.Float64(key, value)
	return err
}

func (err *Error) Floats32(key string, value ) *Error {
    err.e.Floats32(key, value)
	return err
}

func (err *Error) Floats64(key string, value ) *Error {
    err.e.Floats64(key, value)
	return err
}

func (err *Error) Func(value ) *Error {
    err.e.Func(value)
	return err
}

func (err *Error) Hex(key string, value ) *Error {
    err.e.Hex(key, value)
	return err
}

func (err *Error) IPAddr(key string, value IP) *Error {
    err.e.IPAddr(key, value)
	return err
}

func (err *Error) IPPrefix(key string, value IPNet) *Error {
    err.e.IPPrefix(key, value)
	return err
}

func (err *Error) Int(key string, value int) *Error {
    err.e.Int(key, value)
	return err
}

func (err *Error) Int16(key string, value int16) *Error {
    err.e.Int16(key, value)
	return err
}

func (err *Error) Int32(key string, value int32) *Error {
    err.e.Int32(key, value)
	return err
}

func (err *Error) Int64(key string, value int64) *Error {
    err.e.Int64(key, value)
	return err
}

func (err *Error) Int8(key string, value int8) *Error {
    err.e.Int8(key, value)
	return err
}

func (err *Error) Interface(key string, value ) *Error {
    err.e.Interface(key, value)
	return err
}

func (err *Error) Ints(key string, value ) *Error {
    err.e.Ints(key, value)
	return err
}

func (err *Error) Ints16(key string, value ) *Error {
    err.e.Ints16(key, value)
	return err
}

func (err *Error) Ints32(key string, value ) *Error {
    err.e.Ints32(key, value)
	return err
}

func (err *Error) Ints64(key string, value ) *Error {
    err.e.Ints64(key, value)
	return err
}

func (err *Error) Ints8(key string, value ) *Error {
    err.e.Ints8(key, value)
	return err
}

func (err *Error) MACAddr(key string, value HardwareAddr) *Error {
    err.e.MACAddr(key, value)
	return err
}

func (err *Error) Object(key string, value LogObjectMarshaler) *Error {
    err.e.Object(key, value)
	return err
}

func (err *Error) RawCBOR(key string, value ) *Error {
    err.e.RawCBOR(key, value)
	return err
}

func (err *Error) RawJSON(key string, value ) *Error {
    err.e.RawJSON(key, value)
	return err
}

func (err *Error) Stack() *Error {
    err.e.Stack()
	return err
}

func (err *Error) Str(key string, value string) *Error {
    err.e.Str(key, value)
	return err
}

func (err *Error) Stringer(key string, value Stringer) *Error {
    err.e.Stringer(key, value)
	return err
}

func (err *Error) Stringers(key string, value ) *Error {
    err.e.Stringers(key, value)
	return err
}

func (err *Error) Strs(key string, value ) *Error {
    err.e.Strs(key, value)
	return err
}

func (err *Error) Time(key string, value Time) *Error {
    err.e.Time(key, value)
	return err
}

func (err *Error) TimeDiff(key string, value1 Time, value2 Time) *Error {
    err.e.TimeDiff(key, value1, value2)
	return err
}

func (err *Error) Times(key string, value ) *Error {
    err.e.Times(key, value)
	return err
}

func (err *Error) Timestamp() *Error {
    err.e.Timestamp()
	return err
}

func (err *Error) Type(key string, value ) *Error {
    err.e.Type(key, value)
	return err
}

func (err *Error) Uint(key string, value uint) *Error {
    err.e.Uint(key, value)
	return err
}

func (err *Error) Uint16(key string, value uint16) *Error {
    err.e.Uint16(key, value)
	return err
}

func (err *Error) Uint32(key string, value uint32) *Error {
    err.e.Uint32(key, value)
	return err
}

func (err *Error) Uint64(key string, value uint64) *Error {
    err.e.Uint64(key, value)
	return err
}

func (err *Error) Uint8(key string, value uint8) *Error {
    err.e.Uint8(key, value)
	return err
}

func (err *Error) Uints(key string, value ) *Error {
    err.e.Uints(key, value)
	return err
}

func (err *Error) Uints16(key string, value ) *Error {
    err.e.Uints16(key, value)
	return err
}

func (err *Error) Uints32(key string, value ) *Error {
    err.e.Uints32(key, value)
	return err
}

func (err *Error) Uints64(key string, value ) *Error {
    err.e.Uints64(key, value)
	return err
}

func (err *Error) Uints8(key string, value ) *Error {
    err.e.Uints8(key, value)
	return err
}
