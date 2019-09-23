package store

type Local struct {
	vals map[string]string

	getCallback func(string) error
	setCallback func(string, string)
}

func NewLocal() *Local {
	return &Local{
		vals: make(map[string]string),
	}
}

// this method is not thread safe
// TODO could set a mutex lock around setting this
// callback and using this callback
func (l *Local) WithGetCallback(fn func(key string) error) {
	l.getCallback = fn
}

func (l *Local) WithSetCallBack(fn func(key string, val string)) {
	l.setCallback = fn
}

func (l *Local) Get(key string) (val string, retErr error) {
	if l.getCallback != nil {
		defer func() {
			retErr = l.getCallback(key)
		}()
	}
	val, ok := l.vals[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

func (l *Local) Set(key, val string) error {
	if l.setCallback != nil {
		defer l.setCallback(key, val)
	}
	l.vals[key] = val
	return nil
}
