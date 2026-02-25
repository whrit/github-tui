package domain

type Project struct {
	Name string
	URL  string
}

func (p *Project) Key() string {
	return p.Name
}

func (p *Project) Fields() []Field {
	return []Field{
		{Text: p.Name, ColorRole: ColorRoleMuted},
	}
}
