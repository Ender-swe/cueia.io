package routes

import "os"

func browserClerkIssuer() string {
	return os.Getenv("CUEIA_BROWSER_CLERK_ISSUER")
}
