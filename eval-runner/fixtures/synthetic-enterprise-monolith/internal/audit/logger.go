package audit

type Logger struct {
	component string
	events    []string
}

func NewLogger(component string) *Logger {
	return &Logger{component: component}
}

func (l *Logger) Record(event string) {
	l.events = append(l.events, l.component+":"+event)
}

func (l *Logger) Events() []string {
	return append([]string{}, l.events...)
}
