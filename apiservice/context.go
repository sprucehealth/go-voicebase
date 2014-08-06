package apiservice

import (
	"net/http"
	"sync"
	"time"
)

var (
	ctxMu          sync.Mutex
	requestContext = map[*http.Request]*Context{}
)

type CacheKey int

const (
	Patient CacheKey = iota
	PatientId
	PersonId
	Doctor
	DoctorId
	PatientCase
	PatientCaseId
	PatientVisit
	PatientVisitId
	TreatmentPlan
	TreatmentPlanId
	Treatment
	TreatmentId
	RefillRequestId
	RefillRequest
	RequestData
	ERxSource
	FavoriteTreatmentPlan
)

type Context struct {
	AccountId        int64
	Role             string
	RequestStartTime time.Time
	RequestID        int64
	RequestCache     map[CacheKey]interface{}
}

func GetContext(req *http.Request) *Context {
	ctxMu.Lock()
	defer ctxMu.Unlock()
	if ctx := requestContext[req]; ctx != nil {
		return ctx
	}
	ctx := &Context{}
	ctx.RequestCache = make(map[CacheKey]interface{})
	requestContext[req] = ctx
	return ctx
}

func DeleteContext(req *http.Request) {
	ctxMu.Lock()
	delete(requestContext, req)
	ctxMu.Unlock()
}
