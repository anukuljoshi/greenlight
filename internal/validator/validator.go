package validator

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

// function to create Validator instance with empty errors map
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// checks if errors map contains no values
func (v *Validator) Valid() bool {
	return len(v.Errors)==0
}

// add new key value for error if key does not exists
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// adds error message only if validation check is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// helper functions
// check if value in list of strings
func In(value string, list ...string) bool {
	for i := range list {
		if value==list[i] {
			return true
		}
	}
	return false
}

// check if string matches a regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// check if all values in slice of strings are unique
func Unique(values []string) bool {
	var uniqueValues = make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values)==len(uniqueValues)
}
