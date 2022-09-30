package utils

type ConfigScheme struct {
	Port      int64
	Endpoints map[string]string
	Users     map[string]Password
}
