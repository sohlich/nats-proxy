package natsproxy

func (v *Values) Set(key, value string) {

	isSet := false
	for _, v := range v.GetItems() {
		if v.GetKey() == key {
			v.Value = []string{value}
			isSet = true
			break
		}
	}

	if v.Items == nil {
		v.Items = make([]*Value, 0)
	}

	if !isSet {
		v.Items = append(v.Items,
			&Value{Key: &key, Value: []string{value}})
	}

}

func (v *Values) Get(key string) string {

	if v.Items == nil {
		return ""
	}

	var ret []string
	for _, v := range v.GetItems() {
		if v.GetKey() == key {
			ret = v.GetValue()
			break
		}
	}

	if len(ret) == 0 {
		return ""
	}

	return ret[0]
}
