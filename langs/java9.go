package langs

type Java9LangHelper struct {
	BaseHelper
}

// HasBoilerplate returns false as a stub until java9 boilerplate has been implemented
func (lh *Java9LangHelper) HasBoilerplate() bool {
	return false
}
