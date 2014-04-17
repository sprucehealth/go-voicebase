package apiservice

import (
	"carefront/api"
	"carefront/common"
	thriftapi "carefront/thrift/api"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
)

type SignupDoctorHandler struct {
	DataApi api.DataAPI
	AuthApi thriftapi.Auth
}

type DoctorSignedupResponse struct {
	Token    string `json:"token"`
	DoctorId int64  `json:"doctorId, string"`
}

func (d *SignupDoctorHandler) NonAuthenticated() bool {
	return true
}

type SignupDoctorRequestData struct {
	Email        string `schema:"email,required"`
	Password     string `schema:"password,required"`
	FirstName    string `schema:"first_name,required"`
	LastName     string `schema:"last_name,required"`
	Dob          string `schema:"dob,required"`
	Gender       string `schema:"gender,required"`
	ClinicianId  int64  `schema:"clinician_id,required"`
	AddressLine1 string `schema:"address_line_1,required"`
	AddressLine2 string `schema:"address_line_2"`
	City         string `schema:"city"`
	State        string `schema:"state"`
	ZipCode      string `schema:"zip_code"`
	Phone        string `schema:"phone,required"`
}

func (d *SignupDoctorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != HTTP_POST {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse request data: "+err.Error())
		return
	}

	var requestData SignupDoctorRequestData
	if err := schema.NewDecoder().Decode(&requestData, r.Form); err != nil {
		WriteDeveloperError(w, http.StatusBadRequest, "Unable to parse input to signup doctor: "+err.Error())
		return
	}
	// ensure that the date of birth can be correctly parsed
	// Note that the date will be returned as MM/DD/YYYY
	dobParts := strings.Split(requestData.Dob, "/")

	month, err := strconv.Atoi(dobParts[0])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	day, err := strconv.Atoi(dobParts[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(dobParts[2])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// first, create an account for the user
	res, err := d.AuthApi.SignUp(requestData.Email, requestData.Password)
	if _, ok := err.(*thriftapi.LoginAlreadyExists); ok {
		WriteUserError(w, http.StatusBadRequest, "An account with the specified email address already exists.")
		return
	}

	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Internal Servier Error. Unable to register doctor: "+err.Error())
		return
	}

	doctorToRegister := &common.Doctor{
		AccountId:           common.NewObjectId(res.AccountId),
		FirstName:           requestData.FirstName,
		LastName:            requestData.LastName,
		Gender:              requestData.Gender,
		Dob:                 time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
		CellPhone:           requestData.Phone,
		DoseSpotClinicianId: requestData.ClinicianId,
		DoctorAddress: &common.Address{
			AddressLine1: requestData.AddressLine1,
			AddressLine2: requestData.AddressLine2,
			City:         requestData.City,
			State:        requestData.State,
			ZipCode:      requestData.ZipCode,
		},
	}

	// then, register the signed up user as a patient
	doctorId, err := d.DataApi.RegisterDoctor(doctorToRegister)
	if err != nil {
		WriteDeveloperError(w, http.StatusInternalServerError, "Something went wrong when trying to sign up doctor: "+err.Error())
		return
	}

	WriteJSONToHTTPResponseWriter(w, http.StatusOK, DoctorSignedupResponse{res.Token, doctorId})
}
