package domain

import "fmt"

type Issue struct {
	ID        string
	Repo      string
	RepoOwner string
	Number    string
	State     string
	Title     string
	Body      string
	Author    string
	URL       string
	Labels    []Item
	Assignees []Item
	Comments  []Item
	MileStone []Item
	Projects  []Item
}

func (i *Issue) Key() string {
	return i.ID
}

func (i *Issue) Fields() []Field {
	stateRole := ColorRoleSuccess
	if i.State == "CLOSED" {
		stateRole = ColorRoleDanger
	}

	f := []Field{
		{Text: fmt.Sprintf("%s/%s", i.RepoOwner, i.Repo), ColorRole: ColorRoleMuted},
		{Text: i.Number, ColorRole: ColorRoleAccent},
		{Text: i.State, ColorRole: stateRole},
		{Text: i.Author, ColorRole: ColorRoleWarning},
		{Text: i.Title, ColorRole: ColorRoleDefault},
	}

	return f
}
