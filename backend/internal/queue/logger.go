package queue

import (
	"fmt"

	"go.uber.org/zap"
)

// zapAdapter adapts zap.Logger to the asynq.Logger interface.
type zapAdapter struct {
	logger *zap.SugaredLogger
}

func newZapAdapter(logger *zap.Logger) *zapAdapter {
	return &zapAdapter{logger: logger.Sugar()}
}

func (z *zapAdapter) Debug(args ...interface{}) {
	z.logger.Debug(args...)
}

func (z *zapAdapter) Info(args ...interface{}) {
	z.logger.Info(args...)
}

func (z *zapAdapter) Warn(args ...interface{}) {
	z.logger.Warn(args...)
}

func (z *zapAdapter) Error(args ...interface{}) {
	z.logger.Error(args...)
}

func (z *zapAdapter) Fatal(args ...interface{}) {
	z.logger.Fatal(args...)
}

func (z *zapAdapter) Debugf(format string, args ...interface{}) {
	z.logger.Debugf(format, args...)
}

func (z *zapAdapter) Infof(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
}

func (z *zapAdapter) Warnf(format string, args ...interface{}) {
	z.logger.Warnf(format, args...)
}

func (z *zapAdapter) Errorf(format string, args ...interface{}) {
	z.logger.Errorf(format, args...)
}

func (z *zapAdapter) Fatalf(format string, args ...interface{}) {
	z.logger.Fatalf(format, args...)
}

func (z *zapAdapter) Print(args ...interface{}) {
	z.logger.Info(fmt.Sprint(args...))
}

func (z *zapAdapter) Printf(format string, args ...interface{}) {
	z.logger.Infof(format, args...)
}

func (z *zapAdapter) Println(args ...interface{}) {
	z.logger.Info(fmt.Sprint(args...))
}
