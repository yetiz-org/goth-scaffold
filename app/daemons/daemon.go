package daemons

import (
	"fmt"
	"os"
	"reflect"
	"time"

	kkdaemon "github.com/kklab-com/goth-daemon"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/value"
	kkpanic "github.com/kklab-com/goth-panic"
)

type TimerDaemon interface {
	Init() error
	FireLoopAtStart() bool
	Interval() time.Duration
	Loop() error
	LoopStop(sig os.Signal)
	IsStopped() bool
}

type TimerDaemonSetStopped interface {
	SetStopped()
}

func WrapTimerDaemon(daemon TimerDaemon) kkdaemon.Daemon {
	return &_TimerDaemonWrapperStruct{
		TimerDaemon: daemon,
	}
}

type _TimerDaemonWrapperStruct struct {
	TimerDaemon
	stopSig        chan int
	loopStoppedSig chan int
	sig            os.Signal
	state          int
}

func (d *_TimerDaemonWrapperStruct) Registered() error {
	return d.TimerDaemon.Init()
}

func (d *_TimerDaemonWrapperStruct) Start() {
	d.stopSig = make(chan int)
	d.loopStoppedSig = make(chan int)

	if d.FireLoopAtStart() {
		if d.state == 0 {
			d._InvokeLoop()
		}
	}

	go func() {
		for d.state == 0 {
			timer := time.NewTimer(d._TruncateDuration(d.Interval()))
			select {
			case <-timer.C:
				kkpanic.LogCatch(d._InvokeLoop)
				timer.Reset(d.Interval())
			case <-time.After(d.Interval() * 5):
				timer.Reset(d._TruncateDuration(d.Interval()))
				continue
			case <-d.stopSig:
				d.state = 1
				d.SetDaemonStopped()
				d.LoopStop(d.sig)
				close(d.loopStoppedSig)
			}
		}
	}()
}

func (d *_TimerDaemonWrapperStruct) Stop(sig os.Signal) {
	d.sig = sig
	if d.stopSig != nil {
		close(d.stopSig)
	}

	if d.loopStoppedSig != nil {
		<-d.loopStoppedSig
	}
}

func (d *_TimerDaemonWrapperStruct) Restart() {
}

func (d *_TimerDaemonWrapperStruct) Name() string {
	return reflect.TypeOf(d.TimerDaemon).Elem().Name()
}

func (d *_TimerDaemonWrapperStruct) Info() string {
	return value.JsonMarshal(d.TimerDaemon)
}

func (d *_TimerDaemonWrapperStruct) SetDaemonStopped() {
	if daemon, ok := d.TimerDaemon.(TimerDaemonSetStopped); ok {
		daemon.SetStopped()
	}
}

func (d *_TimerDaemonWrapperStruct) _InvokeLoop() {
	if err := d.Loop(); err != nil {
		kklogger.ErrorJ(fmt.Sprintf("TimerDaemon.Loop#%s", d.Name()), err.Error())
	}
}

func (d *_TimerDaemonWrapperStruct) _TruncateDuration(interval time.Duration) time.Duration {
	return time.Now().Truncate(interval).Add(interval).Sub(time.Now())
}

type DefaultTimerDaemon struct {
	stopped bool
}

func (d *DefaultTimerDaemon) Init() error {
	return nil
}

func (d *DefaultTimerDaemon) Interval() time.Duration {
	return time.Minute
}

func (d *DefaultTimerDaemon) FireLoopAtStart() bool {
	return false
}

func (d *DefaultTimerDaemon) Loop() error {
	return nil
}

func (d *DefaultTimerDaemon) LoopStop(sig os.Signal) {
}

func (d *DefaultTimerDaemon) IsStopped() bool {
	return d.stopped
}

func (d *DefaultTimerDaemon) SetStopped() {
	d.stopped = true
}
