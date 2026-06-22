package config

import _ "embed"

var MedicationSuggestionRules []byte

var MedicationDefaultDose []byte

//go:embed medication_suggestion_rules.json
var suggestionRules []byte

//go:embed medication_default_dose.json
var defaultDose []byte

func init() {
	MedicationSuggestionRules = suggestionRules
	MedicationDefaultDose = defaultDose
}
