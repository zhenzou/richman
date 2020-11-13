package richman

type Params struct {
	unmarshal func(interface{}) error
}

func (msg *Params) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *Params) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}
