package base

import "fmt"

// GenTokenUrl ...
func GenTokenUrl(uid string) string {
	return fmt.Sprintf("https://api-user.huami.com/registrations/+86%s/tokens", uid)
}

// GenLoginUrl ...
func GenLoginUrl(uid string) string {
	return "https://account.huami.com/v2/client/login"
}

