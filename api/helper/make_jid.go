package helper

import (
	"strings"

	"go.mau.fi/whatsmeow/types"
)

// reference: https://github.com/asternic/wuzapi/blob/fb110779ada245ab79b4eb17f1462425b8c649e3/wmiau.go#L108
func MakeJID(phone string) (types.JID, bool) {
	if phone == "" {
		return types.NewJID("", types.DefaultUserServer), false
	}

	if phone[0] == '+' {
		phone = phone[1:]
	}

	phoneNumber := strings.Split(phone, "@")[0]
	phoneNumber = strings.Split(phoneNumber, ".")[0]

	for _, c := range phoneNumber {
		if c < '0' || c > '9' {
			recipient, _ := types.ParseJID("")
			return recipient, false
		}
	}

	if !strings.ContainsRune(phone, '@') {
		return types.NewJID(phone, types.DefaultUserServer), true
	}

	recipient, err := types.ParseJID(phone)
	if err != nil {
		return recipient, false
	} else if recipient.User == "" {
		return recipient, false
	}

	return recipient, true
}
