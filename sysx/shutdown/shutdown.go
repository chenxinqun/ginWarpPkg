package shutdown

import (
	"os"
	"os/signal"
	"syscall"
)

type CloseFunc func()

var _ Hook = (*hook)(nil)

// Hook a graceful shutdown hook, default with signals of SIGINT and SIGTERM
type Hook interface {
	// WithSignals add more signals into hook
	WithSignals(signals ...syscall.Signal) Hook

	// Close register shutdown handles
	Close(funcs ...CloseFunc)

	CloseManual()
}

type hook struct {
	ctx   chan os.Signal
	funcs []CloseFunc
}

var shutdownHook Hook

func Set(h Hook) {
	if shutdownHook == nil {
		shutdownHook = h
	}
}

func Default() Hook {
	return shutdownHook
}

// NewHook create a Hook instance
func NewHook() Hook {
	h := &hook{
		ctx: make(chan os.Signal, 1),
	}
	Set(h)
	return h.WithSignals(
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGILL,
		syscall.SIGTRAP,
		syscall.SIGABRT,
		syscall.SIGKILL,
		syscall.SIGTERM,
	)
}

func (h *hook) WithSignals(signals ...syscall.Signal) Hook {
	for _, s := range signals {
		signal.Notify(h.ctx, s)
	}

	return h
}

func (h *hook) Close(funcs ...CloseFunc) {
	h.funcs = funcs
	if <-h.ctx; true {
		signal.Stop(h.ctx)

		for _, f := range funcs {
			f()
		}
	}
}

func (h *hook) CloseManual() {
	for _, f := range h.funcs {
		f()
	}
}
