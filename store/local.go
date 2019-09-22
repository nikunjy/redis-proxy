package store

type Local struct {
	vals map[string]string
}

func NewLocal() *Local {
	return &Local{
		vals: make(map[string]string),
	}
}

func (l *Local) Get(key string) (string, error) {
	val, ok := l.vals[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return val, nil
}

func (l *Local) Put(key, val string) error {
	l.vals[key] = val
	return nil
}
