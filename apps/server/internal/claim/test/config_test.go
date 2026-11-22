package claim_test

import "os"

func testClaimID() string {
	return os.Getenv("CUEIA_TEST_CLAIM_ID")
}
