package gosql

type stringMap map[string]string

func (v stringMap) Keys() []string {
	keys := make([]string, 0, len(v))

	for k := range v {
		keys = append(keys, k)
	}

	return keys
}
