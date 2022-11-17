package base

import (
	"fmt"
	"time"
)

// GenAccessUrl ...
func GenAccessUrl(uid string) string {
	return fmt.Sprintf("https://api-user.huami.com/registrations/%s/tokens", uid)
}

// GenLoginUrl ...
func GenLoginUrl() string {
	return "https://account.huami.com/v2/client/login"
}

// GenSetStepUrl ...
func GenSetStepUrl() string {
	return fmt.Sprintf("https://api-mifit-cn.huami.com/v1/data/band_data.json?&t=%d", time.Now().Unix())
}
