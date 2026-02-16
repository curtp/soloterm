package tag

import "soloterm/shared/validation"

// TagType is the type of tags which are available to the app
type TagType struct {
	Label    string `yaml:"label"`
	Template string `yaml:"template"`
}

func (t *TagType) Validate() *validation.Validator {
	v := validation.NewValidator()

	// label is required
	v.Check("label", len(t.Label) > 0, "cannot be blank")
	// template is required
	v.Check("template", len(t.Template) > 0, "cannot be blank")
	return v
}

func DefaultTagTypes() []TagType {
	return []TagType{
		{Label: "Clock", Template: "[Clock: | ]"},
		{Label: "Event", Template: "[E: | ]"},
		{Label: "Location", Template: "[L: | ]"},
		{Label: "NPC", Template: "[N: | ]"},
		{Label: "Player Character", Template: "[PC: | ]"},
		{Label: "Scene", Template: "[Scene: | ]"},
		{Label: "Thread", Template: "[Thread: | ]"},
		{Label: "Timer", Template: "[Timer: | ]"},
		{Label: "Track", Template: "[Track: | ]"},
	}
}
