package main

import (
	"carefront/libs/erx"
	"strconv"

	"os"
)

func main() {

	clinicId, _ := strconv.ParseInt(os.Getenv("DOSESPOT_CLINIC_ID"), 10, 64)
	userId, _ := strconv.ParseInt(os.Getenv("DOSESPOT_USER_ID"), 10, 64)

	doseSpotService := erx.NewDoseSpotService(clinicId, userId, os.Getenv("DOSESPOT_CLINIC_KEY"), nil)
	err := doseSpotService.GetRefillRequestQueueForClinic()
	if err != nil {
		panic(err.Error())
	}
}
