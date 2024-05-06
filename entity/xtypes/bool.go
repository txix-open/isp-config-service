package xtypes

type Bool struct {
	Value bool
}

func (l Bool) MarshalJSON() ([]byte, error) {
	return l.MarshalText()
}

func (l *Bool) UnmarshalJSON(data []byte) error {
	return l.UnmarshalText(data)
}

func (l Bool) MarshalText() ([]byte, error) {
	if l.Value {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func (l *Bool) UnmarshalText(data []byte) error {
	l.Value = string(data) == "1"
	return nil
}
