package natsproxy

type Variables map[string]*Values

func (h Variables) Get(name string) string {
	if val, ok := h[name]; ok && len(val.Arr) > 0 {
		return val.Arr[0]
	} else {
		return ""
	}
}

func (h Variables) Set(key, val string) {
	h[key] = &Values{[]string{val}}
}

func (h Variables) Add(key, value string) {
	if val, ok := h[key]; ok && len(val.Arr) > 0 {
		val.Arr = append(val.Arr, value)
	} else {
		h.Set(key, value)
	}
}
