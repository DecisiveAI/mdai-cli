package types

type Channels struct {
	messages chan string
	debug    chan string
	errs     chan error
	done     chan struct{}
	task     chan string
}

type Modes struct {
	debug bool
	quiet bool
}

func NewChannels() Channels {
	return Channels{
		messages: make(chan string),
		debug:    make(chan string),
		errs:     make(chan error),
		done:     make(chan struct{}),
		task:     make(chan string),
	}
}

func NewModes(debug, quiet bool) Modes {
	return Modes{
		debug: debug,
		quiet: quiet,
	}
}

func (channels *Channels) Message(message string) {
	channels.messages <- message
}

func (channels *Channels) GetMessage() string {
	return <-channels.messages
}

func (channels *Channels) Debug(message string) {
	channels.debug <- message
}

func (channels *Channels) GetDebug() string {
	return <-channels.debug
}

func (channels *Channels) Error(err error) {
	channels.errs <- err
}

func (channels *Channels) GetError() error {
	return <-channels.errs
}

func (channels *Channels) Done() {
	channels.done <- struct{}{}
}

func (channels *Channels) GetDone() struct{} {
	return <-channels.done
}

func (channels *Channels) Task(task string) {
	channels.task <- task
}

func (channels *Channels) GetTask() string {
	return <-channels.task
}

func (channels *Channels) Close() {
	close(channels.messages)
	close(channels.debug)
	close(channels.errs)
	close(channels.task)
	close(channels.done)
}

func (modes *Modes) Debug() bool {
	return modes.debug
}

func (modes *Modes) Quiet() bool {
	return modes.quiet
}
