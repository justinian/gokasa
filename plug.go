package gokasa

type Plug struct {
	Alias string `json:"alias"`
	Id    string `json:"id"`
	State int    `json:"state"`

	powerStrip *PowerStrip
}

func (p *Plug) IsOn() bool { return p.State != 0 }

func (p *Plug) On() error {
	if p.IsOn() {
		return nil
	}

	return p.powerStrip.setStates([]*Plug{p}, 1)
}

func (p *Plug) Off() error {
	if !p.IsOn() {
		return nil
	}

	return p.powerStrip.setStates([]*Plug{p}, 0)
}
