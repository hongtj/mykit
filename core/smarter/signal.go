package smarter

import (
	"context"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/types"
	"os"
	"os/signal"
	"syscall"
)

func Shutdown() <-chan struct{} {
	return GlobalContext.Done()
}

func RUN(f JOB, tag ...string) {
	fnName := Fn(f).Name()
	if len(tag) > 0 {
		fnName += "->" + tag[0]
	}

	var target JOB = func(ctx context.Context) {
		defer DONE(fnName)
		defer Recover(fnName)

		ADD(fnName)
		f(ctx)
	}

	TracedGo(target,
		fnName,
	)
}

func ADD(name string) {
	DevDebug("ADD " + name)
	GlobalWG.Add(1)
}

func DONE(name string) {
	DevDebug("DONE " + name)
	GlobalWG.Done()
}

func WAIT() {
	GlobalWG.Wait()
}

func handleSignal(sig ...os.Signal) chan os.Signal {
	res := make(chan os.Signal, 1)

	if len(sig) > 0 {
		signal.Notify(res, sig...)

		return res
	}

	sig = []os.Signal{
		os.Interrupt,
		syscall.SIGKILL,
		syscall.SIGTERM,
	}
	signal.Notify(res, sig...)

	return res
}

func TEARDOWN(sig ...os.Signal) {
	<-handleSignal(sig...)
	DevDebug("\nsignal.Notify")

	DevDebug("GlobalCancel")
	GlobalCancel()
	SetStatus(StatusStopped)

	DevDebug("WAIT\n")
	WAIT()

	DevDebug("End of TEARDOWN")
}
