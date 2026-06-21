package logger

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// traceConsoleEncoder 自定义 console 编码器,把 trace_id 字段从结构化输出里抽出来,
// 单独放在 caller 之后、msg 之前的位置。
//
// 输出顺序:
//
//	ts  level  caller  [trace-id]  msg  key=value key=value ...
//
// 没有 trace_id 时输出 [],方便 grep 与对齐。
//
// 实现细节:
//   - ts / level / caller 复用 EncoderConfig 的 EncodeTime/EncodeLevel/EncodeCaller。
//   - 其余 fields 直接追加到 buffer (key=value 形式),不再走 JSON encoder,避免重复输出
//     ts/level/caller/msg,也避免被 Object/Array/Reflected 类型字段触发 noop 丢失。
//   - *buffer.Buffer 不实现完整 PrimitiveArrayEncoder,通过 pae 适配器补齐缺失方法。
type traceConsoleEncoder struct {
	cfg zapcore.EncoderConfig
}

func newTraceConsoleEncoder(cfg zapcore.EncoderConfig) *traceConsoleEncoder {
	return &traceConsoleEncoder{cfg: cfg}
}

func (e *traceConsoleEncoder) Clone() zapcore.Encoder {
	return &traceConsoleEncoder{cfg: e.cfg}
}

func (e *traceConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// 1) 从 fields 中抽出 trace_id
	var traceID string
	var hasTraceID bool
	rest := make([]zapcore.Field, 0, len(fields))
	for _, f := range fields {
		if f.Key == "trace_id" && f.Type == zapcore.StringType {
			traceID = f.String
			hasTraceID = true
			continue
		}
		rest = append(rest, f)
	}

	buf := buffer.NewPool().Get()
	p := pae{Buffer: buf}

	// 2) ts
	if e.cfg.TimeKey != "" && e.cfg.EncodeTime != nil {
		e.cfg.EncodeTime(entry.Time, p)
		buf.AppendByte('\t')
	}
	// 3) level
	if e.cfg.LevelKey != "" && e.cfg.EncodeLevel != nil {
		e.cfg.EncodeLevel(entry.Level, p)
		buf.AppendByte('\t')
	}
	// 4) caller (v1.27 中 EntryCaller 是 struct,Defined 是字段)
	if e.cfg.CallerKey != "" && entry.Caller.Defined {
		e.cfg.EncodeCaller(entry.Caller, p)
	}
	// 5) trace-id (固定位置: caller 之后, msg 之前)
	buf.AppendByte('\t')
	if hasTraceID && traceID != "" {
		buf.AppendByte('[')
		buf.AppendString(traceID)
		buf.AppendByte(']')
	} else {
		buf.AppendString("[]")
	}
	// 6) msg
	if entry.Message != "" && e.cfg.MessageKey != "" {
		buf.AppendByte('\t')
		buf.AppendString(entry.Message)
	}
	// 7) 其余 fields 直接追加 key=value
	for _, f := range rest {
		buf.AppendByte('\t')
		buf.AppendString(f.Key)
		buf.AppendByte('=')
		f.AddTo(p)
	}
	// 8) stacktrace
	if entry.Stack != "" {
		buf.AppendByte('\n')
		buf.AppendString(entry.Stack)
	}
	// 9) 行尾换行,避免日志挤成一行
	buf.AppendByte('\n')

	return buf, nil
}

// ObjectEncoder 接口全部 noop。
// 当前业务只用基本类型字段 (zap.String / Int / Bool / ...),这些字段在 EncodeEntry
// 的 []Field 里被处理,不依赖 ObjectEncoder。Object / Array / Reflected 类型字段会被静默丢弃。
func (e *traceConsoleEncoder) AddArray(k string, v zapcore.ArrayMarshaler) error { return nil }
func (e *traceConsoleEncoder) AddObject(k string, v zapcore.ObjectMarshaler) error {
	return nil
}
func (e *traceConsoleEncoder) AddBinary(k string, v []byte)          {}
func (e *traceConsoleEncoder) AddByteString(k string, v []byte)      {}
func (e *traceConsoleEncoder) AddBool(k string, v bool)              {}
func (e *traceConsoleEncoder) AddComplex128(k string, v complex128)  {}
func (e *traceConsoleEncoder) AddComplex64(k string, v complex64)    {}
func (e *traceConsoleEncoder) AddDuration(k string, v time.Duration) {}
func (e *traceConsoleEncoder) AddFloat64(k string, v float64)        {}
func (e *traceConsoleEncoder) AddFloat32(k string, v float32)        {}
func (e *traceConsoleEncoder) AddInt(k string, v int)                {}
func (e *traceConsoleEncoder) AddInt64(k string, v int64)            {}
func (e *traceConsoleEncoder) AddInt32(k string, v int32)            {}
func (e *traceConsoleEncoder) AddInt16(k string, v int16)            {}
func (e *traceConsoleEncoder) AddInt8(k string, v int8)              {}
func (e *traceConsoleEncoder) AddString(k, v string)                 {}
func (e *traceConsoleEncoder) AddTime(k string, v time.Time)         {}
func (e *traceConsoleEncoder) AddUint(k string, v uint)              {}
func (e *traceConsoleEncoder) AddUint64(k string, v uint64)          {}
func (e *traceConsoleEncoder) AddUint32(k string, v uint32)          {}
func (e *traceConsoleEncoder) AddUint16(k string, v uint16)          {}
func (e *traceConsoleEncoder) AddUint8(k string, v uint8)            {}
func (e *traceConsoleEncoder) AddUintptr(k string, v uintptr)        {}
func (e *traceConsoleEncoder) AddReflected(k string, v interface{}) error {
	return nil
}
func (e *traceConsoleEncoder) OpenNamespace(k string) {}

// ArrayEncoder 接口全部 noop。
func (e *traceConsoleEncoder) AppendArray(v zapcore.ArrayMarshaler) error   { return nil }
func (e *traceConsoleEncoder) AppendObject(v zapcore.ObjectMarshaler) error { return nil }
func (e *traceConsoleEncoder) AppendBool(v bool)                            {}
func (e *traceConsoleEncoder) AppendByteString(v []byte)                    {}
func (e *traceConsoleEncoder) AppendComplex128(v complex128)                {}
func (e *traceConsoleEncoder) AppendComplex64(v complex64)                  {}
func (e *traceConsoleEncoder) AppendDuration(v time.Duration)               {}
func (e *traceConsoleEncoder) AppendTime(v time.Time)                       {}
func (e *traceConsoleEncoder) AppendFloat64(v float64)                      {}
func (e *traceConsoleEncoder) AppendFloat32(v float32)                      {}
func (e *traceConsoleEncoder) AppendInt(v int)                              {}
func (e *traceConsoleEncoder) AppendInt64(v int64)                          {}
func (e *traceConsoleEncoder) AppendInt32(v int32)                          {}
func (e *traceConsoleEncoder) AppendInt16(v int16)                          {}
func (e *traceConsoleEncoder) AppendInt8(v int8)                            {}
func (e *traceConsoleEncoder) AppendString(v string)                        {}
func (e *traceConsoleEncoder) AppendUint(v uint)                            {}
func (e *traceConsoleEncoder) AppendUint64(v uint64)                        {}
func (e *traceConsoleEncoder) AppendUint32(v uint32)                        {}
func (e *traceConsoleEncoder) AppendUint16(v uint16)                        {}
func (e *traceConsoleEncoder) AppendUint8(v uint8)                          {}
func (e *traceConsoleEncoder) AppendUintptr(v uintptr)                      {}
func (e *traceConsoleEncoder) AppendReflected(v interface{}) error          { return nil }

// pae 把 *buffer.Buffer 适配成完整的 PrimitiveArrayEncoder 和 ObjectEncoder。
// 用法:
//   - 作为 zapcore.Field.AddTo(enc ObjectEncoder) 的参数,让字段值写到 buffer。
//   - 作为 zapcore.EncoderConfig.EncodeTime/EncodeLevel/EncodeCaller 的参数。
type pae struct {
	*buffer.Buffer
}

// ObjectEncoder 接口 — 基本类型字段把 value 追加到 buffer。
// 因为在 EncodeEntry 里我们已经手动写了 key= 前缀,这里只追加 value。
func (p pae) AddArray(k string, v zapcore.ArrayMarshaler) error   { return nil }
func (p pae) AddObject(k string, v zapcore.ObjectMarshaler) error { return nil }
func (p pae) AddBinary(k string, v []byte)                        {}
func (p pae) AddByteString(k string, v []byte)                    {}
func (p pae) AddBool(k string, v bool)                            { p.Buffer.AppendBool(v) }
func (p pae) AddComplex128(k string, v complex128) {
	p.Buffer.AppendString(strconv.FormatComplex(v, 'g', -1, 128))
}
func (p pae) AddComplex64(k string, v complex64) {
	p.Buffer.AppendString(strconv.FormatComplex(complex128(v), 'g', -1, 64))
}
func (p pae) AddDuration(k string, v time.Duration) { p.Buffer.AppendInt(v.Nanoseconds()) }
func (p pae) AddFloat64(k string, v float64)        { p.Buffer.AppendFloat(v, 64) }
func (p pae) AddFloat32(k string, v float32)        { p.Buffer.AppendFloat(float64(v), 32) }
func (p pae) AddInt(k string, v int)                { p.Buffer.AppendInt(int64(v)) }
func (p pae) AddInt64(k string, v int64)            { p.Buffer.AppendInt(v) }
func (p pae) AddInt32(k string, v int32)            { p.Buffer.AppendInt(int64(v)) }
func (p pae) AddInt16(k string, v int16)            { p.Buffer.AppendInt(int64(v)) }
func (p pae) AddInt8(k string, v int8)              { p.Buffer.AppendInt(int64(v)) }
func (p pae) AddString(k, v string)                 { p.Buffer.AppendString(v) }
func (p pae) AddTime(k string, v time.Time)         { p.Buffer.AppendString(v.Format(time.RFC3339Nano)) }
func (p pae) AddUint(k string, v uint)              { p.Buffer.AppendUint(uint64(v)) }
func (p pae) AddUint64(k string, v uint64)          { p.Buffer.AppendUint(v) }
func (p pae) AddUint32(k string, v uint32)          { p.Buffer.AppendUint(uint64(v)) }
func (p pae) AddUint16(k string, v uint16)          { p.Buffer.AppendUint(uint64(v)) }
func (p pae) AddUint8(k string, v uint8)            { p.Buffer.AppendUint(uint64(v)) }
func (p pae) AddUintptr(k string, v uintptr)        { p.Buffer.AppendUint(uint64(v)) }
func (p pae) AddReflected(k string, v interface{}) error {
	_, err := fmt.Fprintf(p.Buffer, "%v", v)
	return err
}
func (p pae) OpenNamespace(k string) {}

// ArrayEncoder 接口。
func (p pae) AppendArray(v zapcore.ArrayMarshaler) error   { return nil }
func (p pae) AppendObject(v zapcore.ObjectMarshaler) error { return nil }
func (p pae) AppendDuration(v time.Duration)               {}
func (p pae) AppendTime(v time.Time)                       {}

// buffer.Buffer 已实现 AppendString/AppendBool/AppendInt(int64)/AppendUint(uint64)
// 等方法;这里补全接口要求的其余方法,把类型转换对齐到 buffer.Buffer 的签名。
func (p pae) AppendByteString(b []byte) {
	p.Buffer.AppendBytes(b)
}
func (p pae) AppendComplex128(c complex128) {
	p.Buffer.AppendString(strconv.FormatComplex(c, 'g', -1, 128))
}
func (p pae) AppendComplex64(c complex64) {
	p.Buffer.AppendString(strconv.FormatComplex(complex128(c), 'g', -1, 64))
}
func (p pae) AppendInt(i int)         { p.Buffer.AppendInt(int64(i)) }
func (p pae) AppendInt8(i int8)       { p.Buffer.AppendInt(int64(i)) }
func (p pae) AppendInt16(i int16)     { p.Buffer.AppendInt(int64(i)) }
func (p pae) AppendInt32(i int32)     { p.Buffer.AppendInt(int64(i)) }
func (p pae) AppendInt64(i int64)     { p.Buffer.AppendInt(i) }
func (p pae) AppendUint(i uint)       { p.Buffer.AppendUint(uint64(i)) }
func (p pae) AppendUint8(i uint8)     { p.Buffer.AppendUint(uint64(i)) }
func (p pae) AppendUint16(i uint16)   { p.Buffer.AppendUint(uint64(i)) }
func (p pae) AppendUint32(i uint32)   { p.Buffer.AppendUint(uint64(i)) }
func (p pae) AppendUint64(i uint64)   { p.Buffer.AppendUint(i) }
func (p pae) AppendUintptr(i uintptr) { p.Buffer.AppendUint(uint64(i)) }
func (p pae) AppendFloat32(f float32) { p.Buffer.AppendFloat(float64(f), 32) }
func (p pae) AppendFloat64(f float64) { p.Buffer.AppendFloat(f, 64) }
func (p pae) AppendReflected(v interface{}) error {
	_, err := fmt.Fprintf(p.Buffer, "%v", v)
	return err
}
