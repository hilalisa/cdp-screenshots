package process

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Process struct {
	InitialCounter int
	StartDelay     time.Duration
	Path           string
	Params         []string

	counter      int
	counterMu    sync.RWMutex
	command      *exec.Cmd
	process      *os.Process
	restartingMu sync.Mutex
}

func New(initialCounter int, startDelay time.Duration, path string, params ...string) (*Process, error) {
	fullPath, err := exec.LookPath(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to look up the path")
	}

	if initialCounter <= 0 {
		return nil, errors.New("initial access counter must be greater than 0")
	}

	process := &Process{
		InitialCounter: initialCounter,
		StartDelay:     startDelay,
		Path:           fullPath,
		Params:         params,
	}
	if err := process.Restart(); err != nil {
		return nil, errors.Wrap(err, "unable to run the first start")
	}

	return process, nil
}

func (p *Process) Restart() error {
	if p.command != nil {
		if err := p.command.Process.Kill(); err != nil {
			return errors.Wrap(err, "unable to kill the old process")
		}
	}

	command := exec.Command(p.Path, p.Params...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Start(); err != nil {
		return errors.Wrap(err, "unable to start up the process")
	}

	p.command = command
	p.counter = p.InitialCounter

	time.Sleep(p.StartDelay)
	return nil
}

func (p *Process) Execute(fn func()) {
	p.counterMu.RLock()
	counter := p.counter
	p.counterMu.RUnlock()
	if counter <= 0 {
		p.restartingMu.Lock()

		// Double-check to prevent a race condition
		p.counterMu.RLock()
		counter = p.counter
		p.counterMu.RUnlock()

		if counter <= 0 {
			if err := p.Restart(); err != nil {
				panic(err) // we can panic cuz all this software does depends on it
			}
		}

		p.restartingMu.Unlock()
	}

	fn()

	p.counterMu.Lock()
	p.counter--
	p.counterMu.Unlock()
}
