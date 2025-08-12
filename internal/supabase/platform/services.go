package platform

func GetServices() []Service {
	var svcList []Service

	db := &DatabaseService{}

	svcList = append(svcList, db)

	return svcList
}
