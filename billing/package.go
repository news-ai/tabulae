package billing

func BillingIdToPlanName(plan string) string {
	switch plan {
	case "bronze":
		return "Personal"
	case "silver", "silver-1":
		return "Freelancer"
	case "gold", "gold-1":
		return "Business"
	}

	return "Personal"
}

func UserMaximumSocialAccounts(plan string) int {
	switch plan {
	case "Personal":
		return 20
	case "Freelancer":
		return 500
	case "Business":
		return 100000
	}

	return 0
}

func UserMaximumEmailAccounts(plan string) int {
	switch plan {
	case "Personal":
		return 0
	case "Freelancer":
		return 5
	case "Business":
		return 10
	}

	return 0
}
