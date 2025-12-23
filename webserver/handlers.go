package webserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/oldkingsquid/bg-compiler/processor"
	"github.com/sirupsen/logrus"
)

func compileHandler(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("Action", "CompileHandler")

	bs, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Error reading body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var sub *processor.Submission
	if err := json.Unmarshal(bs, &sub); err != nil {
		logger.Errorf("Error unmarshalling submission: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outputs, err := processor.ProcessSubmission(context.Background(), sub)
	if err != nil {
		logger.WithError(err).Errorf("Error processing submission")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	bs, err = json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		logger.WithError(err).Errorf("Error marshalling outputs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bs)
}
