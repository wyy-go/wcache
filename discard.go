package wcache

// Logger define the logger interface
type Logger interface {
	Errorf(string, ...interface{})
}

// Discard the default logger that will discard all logs of gin-cache
type Discard struct {
}

var _ Logger = (*Discard)(nil)

// NewDiscard a discard logger on which always succeed without doing anything
func NewDiscard() Discard { return Discard{} }

// Debugf implement Logger interface.
func (l Discard) Debugf(string, ...interface{}) {}

// Infof implement Logger interface.
func (l Discard) Infof(string, ...interface{}) {}

// Errorf implement Logger interface.
func (l Discard) Errorf(string, ...interface{}) {}

// Warnf implement Logger interface.
func (l Discard) Warnf(string, ...interface{}) {}

// DPanicf implement Logger interface.
func (l Discard) DPanicf(string, ...interface{}) {}

// Fatalf implement Logger interface.
func (l Discard) Fatalf(string, ...interface{}) {}
