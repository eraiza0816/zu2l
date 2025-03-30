package presenter

import (
	"encoding/json"
	"fmt"
	"io" // Added io import
	"os"
	"time"
	"zutool/internal/models"
)

// JSONPresenter implements the Presenter interface to output data in JSON format.
type JSONPresenter struct {
	// Writer specifies the output destination. Defaults to os.Stdout if nil.
	Writer io.Writer
}

// ensureWriter returns the configured writer or os.Stdout if nil.
func (p *JSONPresenter) ensureWriter() io.Writer {
	if p.Writer == nil {
		return os.Stdout
	}
	return p.Writer
}

// marshalAndPrint marshals the data to indented JSON and prints it to the writer.
func (p *JSONPresenter) marshalAndPrint(data interface{}) error {
	jsonBytes, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}
	_, err = fmt.Fprintln(p.ensureWriter(), string(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}

// PresentPainStatus outputs the pain status data as JSON.
func (p *JSONPresenter) PresentPainStatus(data models.GetPainStatusResponse) error {
	return p.marshalAndPrint(data)
}

// PresentWeatherPoint outputs the weather point data as JSON.
// The kata and keyword parameters are ignored in the JSON output.
func (p *JSONPresenter) PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error {
	return p.marshalAndPrint(data)
}

// PresentWeatherStatus outputs the weather status data as JSON.
// The dayOffset and dayName parameters are ignored in the JSON output.
func (p *JSONPresenter) PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error {
	return p.marshalAndPrint(data)
}

// PresentOtenkiASP outputs the Otenki ASP data as JSON.
// The targetDates, cityName, and cityCode parameters are ignored in the JSON output.
func (p *JSONPresenter) PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error {
	return p.marshalAndPrint(data)
}

// Compile-time check to ensure JSONPresenter implements Presenter.
var _ Presenter = (*JSONPresenter)(nil)
