package domain

type Comment struct {
	ID        string
	Author    string
	UpdatedAt string
	URL       string
	Body      string
}

func (c *Comment) Key() string {
	return c.ID
}

func (c *Comment) Fields() []Field {
	f := []Field{
		{Text: c.Author, ColorRole: ColorRoleWarning},
		{Text: c.UpdatedAt, ColorRole: ColorRoleMuted},
	}

	return f
}
