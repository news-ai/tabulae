package billing

func BillingIdToPlanName(plan string) string {
	switch plan {
	case "bronze": // now "Personal"
		return "Personal"
	case "aluminum": // now "Consultant"
		return "Consultant"
	case "silver", "silver-1": // now "Business"
		return "Freelancer"
	case "gold", "gold-1": // now "Growing Business"
		return "Business"
	}

	return "Personal"
}

func UserMaximumSocialAccounts(plan string) int {
	switch plan {
	case "Personal": // now "Personal"
		return 100
	case "Consultant": // now "Consultant"
		return 250
	case "Freelancer": // now "Business"
		return 500
	case "Business": // now "Growing Business"
		return 100000
	}

	return 0
}

func UserMaximumEmailAccounts(plan string) int {
	switch plan {
	case "Personal": // now "Personal"
		return 0
	case "Consultant": // now "Consultant"
		return 2
	case "Freelancer": // now "Business"
		return 5
	case "Business": // now "Growing Business"
		return 10
	}

	return 0
}
