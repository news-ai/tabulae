package billing

func GetSocialAccount(plan string) int {
	switch plan {
	case "Personal":
		return 20
	case "Business":
		return 500
	case "Ultimate":
		return 100000
	}

	return 0
}
