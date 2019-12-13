package cerror

// Hector Oliveros - 2019
// hector.oliveros.leon@gmail.com

type Severity string

const (
	DEBUG    = Severity("debug")
	INFO     = Severity("info")
	WARNING  = Severity("warning")
	EXPECTED = Severity("expected") // like jwt expired
	ERROR    = Severity("error")
	FATAL    = Severity("fatal")
	SUSPECT  = Severity("suspect")
)
