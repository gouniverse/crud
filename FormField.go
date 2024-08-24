package crud

type FormField struct {
	ID       string
	Type     string
	Name     string
	Value    string
	Label    string
	Help     string
	Options  []FormFieldOption
	OptionsF func() []FormFieldOption
	Required bool
}
