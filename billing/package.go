package billing

func BillingIdToPlanName(plan string) string {
	switch plan {
	case "bronze":
		return "Personal"
	case "silver", "silver-1":
		return "Business"
	case "gold", "gold-1":
		return "Ultimate"
	}

	return "Personal"
}

func UserMaximumSocialAccounts(plan string) int {
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

func UserMaximumEmailAccounts(plan string) int {
	switch plan {
	case "Personal":
		return 1
	case "Business":
		return 5
	case "Ultimate":
		return 10
	}

	return 0
}

