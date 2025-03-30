package presenter

import (
	"time" // time.Time を使うためインポート
	"zutool/internal/models"
)

// Presenter defines the interface for displaying API results.
// Implementations will handle different output formats (e.g., table, JSON).
type Presenter interface {
	// PresentPainStatus displays the pain status forecast.
	PresentPainStatus(data models.GetPainStatusResponse) error

	// PresentWeatherPoint displays the weather point search results.
	// kata: whether to include katakana name.
	// keyword: the search keyword used.
	PresentWeatherPoint(data models.GetWeatherPointResponse, kata bool, keyword string) error

	// PresentWeatherStatus displays the detailed weather status.
	// dayOffset: the offset from today (-1, 0, 1, 2).
	// dayName: the name of the day ("yesterday", "today", etc.).
	PresentWeatherStatus(data models.GetWeatherStatusResponse, dayOffset int, dayName string) error

	// PresentOtenkiASP displays the Otenki ASP weather information.
	// targetDates: the specific dates to display.
	// cityName: the name of the city.
	// cityCode: the code of the city.
	PresentOtenkiASP(data models.GetOtenkiASPResponse, targetDates []time.Time, cityName, cityCode string) error
}
