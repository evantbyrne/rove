package rove

import "github.com/evantbyrne/trance"

type PreferenceType string

const DefaultMachine PreferenceType = "default_machine_name"

func GetPreference(preferenceType PreferenceType) string {
	value := ""
	trance.Query[Preference]().
		Filter("name", "=", string(preferenceType)).
		First().
		Then(func(preference *Preference) error {
			value = preference.Value
			return nil
		})
	return value
}

func SetPreference(preferenceType PreferenceType, value string) *trance.QueryResultStreamer[Preference] {
	_, err := trance.Query[Preference]().
		Filter("name", "=", string(preferenceType)).
		CollectFirst()

	if err == nil {
		return trance.Query[Preference]().
			Filter("name", "=", string(preferenceType)).
			UpdateMap(map[string]any{"value": value})
	}
	return trance.Query[Preference]().InsertMap(map[string]any{
		"name":  string(preferenceType),
		"value": value,
	})
}
