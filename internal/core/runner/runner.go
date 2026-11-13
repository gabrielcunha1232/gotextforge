package runner

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

const maxLines = 5000

var ansi = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]|\x1b\][^\x07\x1b]*(\x07|\x1b\\)|\x1b[()][A-Za-z0-9]`)

func CleanLine(s string) string {
	s = strings.TrimSuffix(s, "\r")
	if i := strings.LastIndex(s, "\r"); i >= 0 {
		s = s[i+1:]
	}
	return ansi.ReplaceAllString(s, "")
}

type State int

const (
	Running State = iota
	Succeeded
	Failed
	Canceled
)

type Snapshot struct {
	Lines    []string
	Total    int
	State    State
	ExitCode int
	Err      error
	Elapsed  time.Duration
}

type Job struct {
	Command string
	Dir     string

	mu       sync.Mutex
	lines    []string
	dropped  int
	state    State
	exitCode int
	err      error
	started  time.Time
	finished time.Time
	canceled bool
	proc     *exec.Cmd
	notify   func(final bool)
	dirty    bool
}

func Start(command, dir string, notify func(final bool)) (*Job, error) {
	if strings.TrimSpace(command) == "" {
		return nil, errors.New("comando vazio")
	}
	j := &Job{Command: command, Dir: dir, notify: notify, started: time.Now()}

	cmd := shellCommand(command)
	cmd.Dir = dir
	cmd.Stdin = nil
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	prepare(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	j.proc = cmd

	var wg sync.WaitGroup
	wg.Add(2)
	go j.read(stdout, &wg)
	go j.read(stderr, &wg)

	stopTicker := make(chan struct{})
	go j.tick(stopTicker)

	go func() {
		wg.Wait()
		waitErr := cmd.Wait()
		close(stopTicker)

		j.mu.Lock()
		j.finished = time.Now()
		switch {
		case j.canceled:
			j.state = Canceled
		case waitErr == nil:
			j.state = Succeeded
		default:
			j.state = Failed
			var exit *exec.ExitError
			if errors.As(waitErr, &exit) {
				j.exitCode = exit.ExitCode()
			} else {
				j.err = waitErr
			}
		}
		j.mu.Unlock()
		if j.notify != nil {
			j.notify(true)
		}
	}()
	return j, nil
}

func (j *Job) read(r io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	br := bufio.NewReaderSize(r, 64*1024)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			j.append(CleanLine(strings.TrimSuffix(line, "\n")))
		}
		if err != nil {
			return
		}
	}
}

func (j *Job) append(line string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.lines = append(j.lines, line)
	if len(j.lines) > maxLines {
		cut := len(j.lines) - maxLines
		j.lines = append([]string(nil), j.lines[cut:]...)
		j.dropped += cut
	}
	j.dirty = true
}

func (j *Job) tick(stop <-chan struct{}) {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			j.mu.Lock()
			dirty := j.dirty
			j.dirty = false
			j.mu.Unlock()
			if dirty && j.notify != nil {
				j.notify(false)
			}
		}
	}
}

func (j *Job) Snapshot() Snapshot {
	j.mu.Lock()
	defer j.mu.Unlock()
	end := j.finished
	if j.state == Running {
		end = time.Now()
	}
	return Snapshot{
		Lines:    append([]string(nil), j.lines...),
		Total:    j.dropped + len(j.lines),
		State:    j.state,
		ExitCode: j.exitCode,
		Err:      j.err,
		Elapsed:  end.Sub(j.started),
	}
}

func (j *Job) Cancel() {
	j.mu.Lock()
	if j.state != Running || j.proc == nil || j.proc.Process == nil {
		j.mu.Unlock()
		return
	}
	j.canceled = true
	pid := j.proc.Process.Pid
	j.mu.Unlock()
	killTree(j.proc, pid)
}

func (j *Job) Wait(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if j.Snapshot().State != Running {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}
