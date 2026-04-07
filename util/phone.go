package util

func HidePhone(phone string) string {
	length := len(phone)
	if length > 8 {
		return phone[0:length-8] + "****" + phone[length-4:]
	}
	return "****" + phone[length-4:]

}
