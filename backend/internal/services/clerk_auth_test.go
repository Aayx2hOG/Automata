package services

import (
	clerk "github.com/clerk/clerk-sdk-go/v2"
	"testing"
)

func TestVerifiedPrimaryEmail(t *testing.T) {
	for _, tc := range []struct {
		name                              string
		verified, banned, locked, primary bool
		ok                                bool
	}{
		{"verified", true, false, false, true, true},
		{"unverified", false, false, false, true, false},
		{"banned", true, true, false, true, false},
		{"locked", true, false, true, true, false},
		{"not primary", true, false, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			profile := &clerk.User{Banned: tc.banned, Locked: tc.locked}
			if tc.primary {
				profile.PrimaryEmailAddressID = clerk.String("email_1")
			}
			status := "unverified"
			if tc.verified {
				status = "verified"
			}
			profile.EmailAddresses = []*clerk.EmailAddress{{ID: "email_1", EmailAddress: "test@example.com", Verification: &clerk.Verification{Status: status}}}
			email, err := verifiedPrimaryEmail(profile)
			if (err == nil) != tc.ok {
				t.Fatalf("unexpected result: %v", err)
			}
			if tc.ok && email != "test@example.com" {
				t.Fatalf("unexpected email %s", email)
			}
		})
	}
}
