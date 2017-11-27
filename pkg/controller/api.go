package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// API Paths
	APIPathLaunchContainer = "/launchContainer"
	APIPathCallService     = "/service"
	APIPathAddNode         = "/node"
	APIPathHealth          = "/health"

	// QueryString Parameters
	QStringUserID  = "userID"
	QStringProject = "project"
	QStringMethod  = "method"
)

type API struct {
	logger     log.Logger
	controller ControllerService
	duration   *prometheus.HistogramVec
}

func NewAPI(
	logger log.Logger,
	ctrl ControllerService,
	duration *prometheus.HistogramVec,
) *API {
	return &API{
		logger:     logger,
		controller: ctrl,
		duration:   duration,
	}
}

type interceptingWriter struct {
	code int
	http.ResponseWriter
}

func (iw *interceptingWriter) WriteHeader(code int) {
	iw.code = code
	iw.ResponseWriter.WriteHeader(code)
}

func (iw *interceptingWriter) Flush() {
	if f, ok := iw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (a *API) Close() error {
	return nil
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	iw := &interceptingWriter{http.StatusOK, w}
	w = iw
	defer func(begin time.Time) {
		a.duration.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(iw.code),
		).Observe(time.Since(begin).Seconds())
	}(time.Now())

	switch {
	case r.Method == "POST" && r.URL.Path == APIPathLaunchContainer:
		a.launchProject(w, r)
	case r.Method == "POST" && r.URL.Path == APIPathCallService:
		a.callService(w, r)
	case r.Method == "GET" && r.URL.Path == APIPathHealth:
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

func (a *API) callService(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	userID := queryValues.Get(QStringUserID)
	project := queryValues.Get(QStringProject)
	method := queryValues.Get(QStringMethod)

	p := Project{
		UserID:      userID,
		ProjectName: project,
	}

	svcURL, err := a.controller.CallService(p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, svcURL+"/"+method, http.StatusSeeOther)
}

func (a *API) launchProject(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var ci ContainerInfo
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&ci); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.controller.LaunchContainer(ci)
}
