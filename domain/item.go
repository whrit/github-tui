package domain

// ColorRole is a semantic color identifier resolved by the UI theme layer.
type ColorRole string

const (
	ColorRoleDefault ColorRole = "default"
	ColorRoleMuted   ColorRole = "muted"
	ColorRoleAccent  ColorRole = "accent"
	ColorRoleSuccess ColorRole = "success"
	ColorRoleDanger  ColorRole = "danger"
	ColorRoleWarning ColorRole = "warning"
)

type Item interface {
	Key() string
	Fields() []Field
}

type Field struct {
	Text      string
	ColorRole ColorRole
}
