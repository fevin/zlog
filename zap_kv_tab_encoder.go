package zlog

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	_bufferPool   = buffer.NewPool()
	_zapKVTabPool = sync.Pool{New: func() interface{} {
		return &zapKVTabEncoder{}
	}}
)

func getZapKVTabEncoder() *zapKVTabEncoder {
	return _zapKVTabPool.Get().(*zapKVTabEncoder)
}

func putZapKVTabEncoder(enc *zapKVTabEncoder) {
	if enc.reflectBuf != nil {
		enc.reflectBuf.Free()
	}
	enc.EncoderConfig = nil
	enc.buf = nil
	enc.reflectBuf = nil
	enc.reflectEnc = nil
	_zapKVTabPool.Put(enc)
}

func newZapKVTabEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &zapKVTabEncoder{
		EncoderConfig: &cfg,
		buf:           _bufferPool.Get(),
	}
}

type zapKVTabEncoder struct {
	*zapcore.EncoderConfig
	buf *buffer.Buffer

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc *json.Encoder
}

func (enc *zapKVTabEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *zapKVTabEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *zapKVTabEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *zapKVTabEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *zapKVTabEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *zapKVTabEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *zapKVTabEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *zapKVTabEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *zapKVTabEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *zapKVTabEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = _bufferPool.Get()
		enc.reflectEnc = json.NewEncoder(enc.reflectBuf)

		// For consistency with our custom JSON encoder.
		enc.reflectEnc.SetEscapeHTML(false)
	} else {
		enc.reflectBuf.Reset()
	}
}

var nullLiteralBytes = []byte("null")

// Only invoke the standard JSON encoder if there is actually something to
// encode; otherwise write JSON null literal directly.
func (enc *zapKVTabEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	enc.resetReflectBuf()
	if err := enc.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	enc.reflectBuf.TrimNewline()
	return enc.reflectBuf.Bytes(), nil
}

func (enc *zapKVTabEncoder) AddReflected(key string, obj interface{}) error {
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *zapKVTabEncoder) OpenNamespace(_ string) {
}

func (enc *zapKVTabEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

func (enc *zapKVTabEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

func (enc *zapKVTabEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

func (enc *zapKVTabEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}

func (enc *zapKVTabEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	return err
}

func (enc *zapKVTabEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *zapKVTabEncoder) AppendByteString(val []byte) {
	enc.buf.Write(val)
}

func (enc *zapKVTabEncoder) AppendComplex128(val complex128) {
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
}

func (enc *zapKVTabEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	enc.EncodeDuration(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *zapKVTabEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *zapKVTabEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *zapKVTabEncoder) AppendString(val string) {
	enc.buf.AppendString(val)
}

func (enc *zapKVTabEncoder) AppendTimeLayout(time time.Time, layout string) {
	enc.buf.AppendTime(time, layout)
}

func (enc *zapKVTabEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	enc.EncodeTime(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *zapKVTabEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *zapKVTabEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *zapKVTabEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *zapKVTabEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *zapKVTabEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *zapKVTabEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *zapKVTabEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *zapKVTabEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *zapKVTabEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *zapKVTabEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *zapKVTabEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *zapKVTabEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *zapKVTabEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *zapKVTabEncoder) AppendFloat64(v float64)            { enc.appendFloat(v, 64) }
func (enc *zapKVTabEncoder) AppendFloat32(v float32)            { enc.appendFloat(float64(v), 32) }
func (enc *zapKVTabEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *zapKVTabEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *zapKVTabEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *zapKVTabEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *zapKVTabEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *zapKVTabEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *zapKVTabEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *zapKVTabEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *zapKVTabEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *zapKVTabEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *zapKVTabEncoder) clone() *zapKVTabEncoder {
	clone := getZapKVTabEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.buf = _bufferPool.Get()
	return clone
}

func (enc *zapKVTabEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	// 2015-12-02T00:00:07.099+0800
	final.buf.AppendString(enc.TimeKey)
	final.buf.AppendByte('=')
	final.EncodeTime(ent.Time, final)
	final.buf.AppendByte('\t')

	// foo.go:123
	if final.CallerKey != "" && ent.Caller.Defined {
		final.buf.AppendString(enc.CallerKey)
		final.buf.AppendByte('=')
		final.EncodeCaller(ent.Caller, final)
		final.buf.AppendByte('\t')
	}

	if final.MessageKey != "" {
		final.buf.AppendString(enc.MessageKey)
		final.buf.AppendByte('=')
		final.buf.AppendString(ent.Message)
	}

	// white space required!!
	final.buf.AppendByte('\t')

	// default dltag
	for i := range fields {
		fields[i].AddTo(final)
	}
	final.buf.AppendString(enc.LineEnding)

	ret := final.buf
	putZapKVTabEncoder(final)
	return ret, nil
}

func (enc *zapKVTabEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *zapKVTabEncoder) addKey(key string) {
	enc.addElementSeparator()
	enc.buf.AppendString(key)
	enc.buf.AppendByte('=')
}

func (enc *zapKVTabEncoder) addElementSeparator() {
	enc.buf.AppendByte('\t')
}

func (enc *zapKVTabEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}
