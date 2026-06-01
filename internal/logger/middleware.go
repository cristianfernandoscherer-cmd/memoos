package logger

import (
	"time"

	"go.uber.org/zap"
)

type OperationLogger struct {
	logger *Logger
	fields []interface{}
	start  time.Time
	name   string
}

func NewOperation(logger *Logger, name string, fields ...interface{}) *OperationLogger {
	op := &OperationLogger{
		logger: logger,
		fields: fields,
		start:  time.Now(),
		name:   name,
	}
	op.logger.Debugf("operation started: %s %v", name, fields)
	return op
}

func (op *OperationLogger) Success() {
	duration := time.Since(op.start)
	op.logger.Infow("operation succeeded",
		append(op.fields, "operation", op.name, "duration_ms", duration.Milliseconds())...,
	)
}

func (op *OperationLogger) SuccessWith(fields ...interface{}) {
	duration := time.Since(op.start)
	op.logger.Infow("operation succeeded",
		append(append(op.fields, fields...), "operation", op.name, "duration_ms", duration.Milliseconds())...,
	)
}

func (op *OperationLogger) Error(err error) {
	duration := time.Since(op.start)
	op.logger.Errorw("operation failed",
		append(op.fields, "operation", op.name, "duration_ms", duration.Milliseconds(), "error", err.Error())...,
	)
}

func (op *OperationLogger) ErrorWith(err error, fields ...interface{}) {
	duration := time.Since(op.start)
	op.logger.Errorw("operation failed",
		append(append(op.fields, fields...), "operation", op.name, "duration_ms", duration.Milliseconds(), "error", err.Error())...,
	)
}

type LogField struct {
	Key   string
	Value interface{}
}

func F(key string, value interface{}) LogField {
	return LogField{Key: key, Value: value}
}

func LogFields(fields ...LogField) []interface{} {
	result := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		result = append(result, f.Key, f.Value)
	}
	return result
}

func KV(key string, value interface{}) interface{} {
	return zap.String(key, toString(value))
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
