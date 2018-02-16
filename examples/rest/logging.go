package main

import (
	"sync"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

var (
	bufferpool  = buffer.NewPool()
	encoderpool = sync.Pool{New: func() interface{} {
		return &RequestEncoder{}
	}}
)

type RequestEncoder struct {
	*zapcore.EncoderConfig
	buf *buffer.Buffer
}

func NewRequestEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	encoder := encoderpool.Get().(*RequestEncoder)
	encoder.EncoderConfig = &cfg
	encoder.buf = bufferpool.Get()

	return encoder
}

func (e *RequestEncoder) free() {
	e.EncoderConfig = nil
	e.buf.Free()
	e.buf = nil
	encoderpool.Put(e)
}

func (e *RequestEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error {
	return nil
}

func (e *RequestEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error {
	return nil
}

func (e *RequestEncoder) AddBinary(key string, value []byte) {

}

func (e *RequestEncoder) AddByteString(key string, value []byte) {

}

func (e *RequestEncoder) AddBool(key string, value bool) {

}

func (e *RequestEncoder) AddComplex128(key string, value complex128) {

}

func (e *RequestEncoder) AddComplex64(key string, value complex64) {

}

func (e *RequestEncoder) AddDuration(key string, value time.Duration) {

}

func (e *RequestEncoder) AddFloat64(key string, value float64) {

}

func (e *RequestEncoder) AddFloat32(key string, value float32) {

}

func (e *RequestEncoder) AddInt(key string, value int) {

}

func (e *RequestEncoder) AddInt64(key string, value int64) {

}

func (e *RequestEncoder) AddInt32(key string, value int32) {

}

func (e *RequestEncoder) AddInt16(key string, value int16) {

}

func (e *RequestEncoder) AddInt8(key string, value int8) {

}

func (e *RequestEncoder) AddString(key, value string) {

}

func (e *RequestEncoder) AddTime(key string, value time.Time) {

}

func (e *RequestEncoder) AddUint(key string, value uint) {

}

func (e *RequestEncoder) AddUint64(key string, value uint64) {

}

func (e *RequestEncoder) AddUint32(key string, value uint32) {

}

func (e *RequestEncoder) AddUint16(key string, value uint16) {

}

func (e *RequestEncoder) AddUint8(key string, value uint8) {

}

func (e *RequestEncoder) AddUintptr(key string, value uintptr) {

}

func (e *RequestEncoder) AddReflected(key string, value interface{}) error {
	return nil
}

func (e *RequestEncoder) OpenNamespace(key string) {

}

func (e *RequestEncoder) Clone() zapcore.Encoder {
	clone := encoderpool.Get().(*RequestEncoder)
	clone.EncoderConfig = e.EncoderConfig
	clone.buf = bufferpool.Get()
	clone.buf.Write(e.buf.Bytes())

	return clone
}

func (e *RequestEncoder) EncodeEntry(zapcore.Entry, []zapcore.Field) (*buffer.Buffer, error) {

	return nil, nil
}
