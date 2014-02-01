package erx

import (
	"carefront/common"
)

type ERxAPI interface {
	GetDrugNamesForDoctor(prefix string) ([]string, error)
	GetDrugNamesForPatient(prefix string) ([]string, error)
	SearchForMedicationStrength(medicationName string) ([]string, error)
	SelectMedication(medicationName, medicationStrength string) (medication *Medication, err error)
	StartPrescribingPatient(Patient *common.Patient, Treatments []*common.Treatment) error
	SendMultiplePrescriptions(Patient *common.Patient, Treatments []*common.Treatment) ([]int64, error)
}

type Medication struct {
	DrugDBIds               map[string]string
	OTC                     bool
	DispenseUnitId          int
	DispenseUnitDescription string
}
