package permissions

func AccessToObject(objectId int64, userId int64) bool {
	if objectId != userId {
		return false
	}
	return true
}
