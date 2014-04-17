package apiservice

import (
	"carefront/common"
	"carefront/libs/address_validation"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// these are valid area codes as obtained from dosespot, for what they claim is a list last updated from January 2011
var (
	validAreaCodes = []string{"201", "202", "203", "204", "205", "206", "207", "208", "209", "210", "212", "213", "214", "215", "216",
		"217", "218", "219", "224", "225", "226", "228", "229", "231", "234", "239", "240", "242", "246", "248", "250", "251", "252", "253",
		"254", "256", "260", "262", "264", "267", "268", "269", "270", "276", "281", "284", "289", "301", "302", "303", "304", "305", "306",
		"307", "308", "309", "310", "312", "313", "314", "315", "316", "317", "318", "319", "320", "321", "323", "325", "330", "331", "334",
		"336", "337", "339", "340", "343", "345", "347", "351", "352", "360", "361", "385", "386", "401", "402", "403", "404", "405", "406",
		"407", "408", "409", "410", "412", "413", "414", "415", "416", "417", "418", "419", "423", "424", "425", "430", "432", "434", "435",
		"438", "440", "441", "442", "443", "450", "456", "458", "469", "470", "473", "475", "478", "479", "480", "484", "500", "501", "502",
		"503", "504", "505", "506", "507", "508", "509", "510", "512", "513", "514", "515", "516", "517", "518", "519", "520", "530", "533",
		"534", "540", "541", "551", "559", "561", "562", "563", "567", "570", "571", "573", "574", "575", "579", "580", "581", "585", "586",
		"587", "600", "601", "602", "603", "604", "605", "606", "607", "608", "609", "610", "612", "613", "614", "615", "616", "617", "618",
		"619", "620", "623", "626", "630", "631", "636", "641", "646", "647", "649", "650", "651", "657", "660", "661", "662", "664", "670",
		"671", "678", "681", "682", "684", "700", "701", "702", "703", "704", "705", "706", "707", "708", "709", "710", "712", "713", "714",
		"715", "716", "717", "718", "719", "720", "724", "727", "731", "732", "734", "740", "747", "754", "757", "758", "760", "762", "763",
		"765", "767", "769", "770", "772", "773", "774", "775", "778", "779", "780", "781", "784", "785", "786", "787", "800", "801", "802",
		"803", "804", "805", "806", "807", "808", "809", "810", "812", "813", "814", "815", "816", "817", "818", "819", "828", "829", "830",
		"831", "832", "843", "845", "847", "848", "849", "850", "855", "856", "857", "858", "859", "860", "862", "863", "864", "865", "866",
		"867", "868", "869", "870", "872", "876", "877", "878", "888", "900", "901", "902", "903", "904", "905", "906", "907", "908", "909",
		"910", "912", "913", "914", "915", "916", "917", "918", "919", "920", "925", "928", "931", "936", "937", "938", "939", "940", "941",
		"947", "949", "951", "952", "954", "956", "970", "971", "972", "973", "978", "979", "980", "985", "989"}
)

func (d *DoctorPatientUpdateHandler) validatePatientInformationAccordingToSurescriptsRequirements(patient *common.Patient, addressValidationApi address_validation.AddressValidationAPI) error {

	if patient.FirstName == "" {
		return errors.New("First name is required")
	}

	if patient.LastName == "" {
		return errors.New("Last name is required")
	}

	if patient.Dob.Month == 0 || patient.Dob.Year == 0 || patient.Dob.Day == 0 {
		return errors.New("Dob is invalid. Required format is YYYY-MM-DD")
	}

	if len(patient.PhoneNumbers) == 0 {
		return errors.New("Atleast one phone number is required")
	}

	if patient.PatientAddress.AddressLine1 == "" {
		return errors.New("AddressLine1 of address is required")
	}

	if patient.PatientAddress.City == "" {
		return errors.New("City in address is required")
	}

	if patient.PatientAddress.State == "" {
		return errors.New("State in address is required")
	}

	// following field lengths are surescripts requirements
	longFieldLength := 35
	shortFieldLength := 10
	phoneNumberLength := 25

	if len(patient.Prefix) > shortFieldLength {
		return fmt.Errorf("Prefix cannot be longer than %d characters in length", shortFieldLength)
	}

	if len(patient.Suffix) > shortFieldLength {
		return fmt.Errorf("Suffix cannot be longer than %d characters in length", shortFieldLength)
	}

	if len(patient.FirstName) > longFieldLength {
		return fmt.Errorf("First name cannot be longer than %d characters", longFieldLength)
	}

	if len(patient.MiddleName) > longFieldLength {
		return fmt.Errorf("Middle name cannot be longer than %d characters", longFieldLength)
	}

	if len(patient.LastName) > longFieldLength {
		return fmt.Errorf("Last name cannot be longer than %d characters", longFieldLength)
	}

	if len(patient.PatientAddress.AddressLine1) > longFieldLength {
		return fmt.Errorf("AddressLine1 of patient address cannot be longer than %d characters", longFieldLength)
	}

	if len(patient.PatientAddress.AddressLine2) > longFieldLength {
		return fmt.Errorf("AddressLine2 of patient address cannot be longer than %d characters", longFieldLength)
	}

	if len(patient.PatientAddress.City) > longFieldLength {
		return fmt.Errorf("City cannot be longer than %d characters", longFieldLength)
	}

	for _, phoneNumber := range patient.PhoneNumbers {
		if len(phoneNumber.Phone) > 25 {
			return fmt.Errorf("Phone numbers cannot be longer than %d digits", phoneNumberLength)
		}
	}

	if err := validateAddress(d.DataApi, patient.PatientAddress, addressValidationApi); err != nil {
		return err
	}

	for _, phoneNumber := range patient.PhoneNumbers {
		if err := validatePhoneNumber(phoneNumber.Phone); err != nil {
			return err
		}
	}

	return nil
}

func trimSpacesFromPatientFields(patient *common.Patient) {
	patient.FirstName = strings.TrimSpace(patient.FirstName)
	patient.LastName = strings.TrimSpace(patient.LastName)
	patient.MiddleName = strings.TrimSpace(patient.MiddleName)
	patient.Suffix = strings.TrimSpace(patient.Suffix)
	patient.Prefix = strings.TrimSpace(patient.Prefix)
	patient.PatientAddress.AddressLine1 = strings.TrimSpace(patient.PatientAddress.AddressLine1)
	patient.PatientAddress.AddressLine2 = strings.TrimSpace(patient.PatientAddress.AddressLine2)
	patient.PatientAddress.City = strings.TrimSpace(patient.PatientAddress.City)
	patient.PatientAddress.State = strings.TrimSpace(patient.PatientAddress.State)
}

func validatePhoneNumber(phoneNumber string) error {
	// phone number has to be 10 digits long
	if len(phoneNumber) < 10 {
		return fmt.Errorf("Invalid phone number")
	}

	// if first 10 characteres are not digits, phone number is not valid
	_, err := strconv.ParseInt(phoneNumber[0:10], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid phone number")
	}

	if !isValidAreaCode(phoneNumber[0:3]) {
		return fmt.Errorf("Invalid area code")
	}

	if len(phoneNumber) > 10 {
		// only acceptable character for extension is x
		if phoneNumber[10] != 'x' && phoneNumber[10] != 'X' {
			return fmt.Errorf("Invalid extension for phone number. Extension must to start with an 'x'")
		}

		if len(phoneNumber) == 11 {
			return fmt.Errorf("Invalid extension for phone number. 'x' must follow the extension")
		}

		_, err := strconv.ParseInt(phoneNumber[11:], 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid extension for phone number. Extension can only be digits")
		}
	}

	return nil
}

func isValidAreaCode(areaCode string) bool {
	for _, validAreaCode := range validAreaCodes {
		if validAreaCode == areaCode {
			return true
		}
	}
	return false
}
