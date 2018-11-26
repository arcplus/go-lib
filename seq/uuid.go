package seq

import "github.com/google/uuid"

// UUID returns random uuid
func UUID() string {
	u, err := uuid.NewRandom()
	if err == nil {
		return u.String()
	}
	return ""
}
