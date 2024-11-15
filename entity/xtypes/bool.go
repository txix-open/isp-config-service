package xtypes

type Bool bool

func (l Bool) MarshalJSON() ([]byte, error) {
	return l.MarshalText()
}

func (l *Bool) UnmarshalJSON(data []byte) error {
	return l.UnmarshalText(data)
}

func (l Bool) MarshalText() ([]byte, error) {
	if l {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func (l *Bool) UnmarshalText(data []byte) error {
	*l = string(data) == "1"
	return nil
}
