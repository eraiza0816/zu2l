package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eraiza0816/zu2l/api"
)

func main() {
	// APIクライアントを作成
	client := api.NewClient("", "", 10*time.Second) // デフォルト設定を使用

	// 例: 東京の天気ステータスを取得
	cityCode := "13101" // 東京
	fmt.Printf("Fetching weather status for city code: %s\n", cityCode)
	weatherStatus, err := client.GetWeatherStatus(cityCode)
	if err != nil {
		log.Fatalf("Error getting weather status: %v", err)
	}

	fmt.Printf("Weather Status for %s (%s):\n", weatherStatus.PlaceName, cityCode)
	fmt.Println("------------------------------------")
	if len(weatherStatus.Today) > 0 {
		fmt.Println("Today:")
		for _, status := range weatherStatus.Today {
			fmt.Printf("  Time: %s, Temp: %.1f°C, Pressure: %.1fhPa, Weather: %s\n",
				status.DateTime.Format("15:04"), status.Temperature, status.Pressure, status.Weather)
		}
	} else {
		fmt.Println("  No data for today.")
	}
	fmt.Println("------------------------------------")

	// 例: "渋谷"で地点を検索
	keyword := "渋谷"
	fmt.Printf("\nSearching for weather point with keyword: %s\n", keyword)
	weatherPoint, err := client.GetWeatherPoint(keyword)
	if err != nil {
		log.Fatalf("Error getting weather point: %v", err)
	}

	fmt.Printf("Weather Points found for '%s':\n", keyword)
	fmt.Println("------------------------------------")
	if len(weatherPoint.Result.Root) > 0 {
		for _, point := range weatherPoint.Result.Root {
			fmt.Printf("  City Code: %s, Name: %s (%s)\n", point.CityCode, point.Name, point.Kana)
		}
	} else {
		fmt.Println("  No points found.")
	}
	fmt.Println("------------------------------------")

	// 例: 北海道の痛み予報を取得 (地点コード 01101 を設定)
	areaCode := "01" // 北海道
	setPoint := "01101" // 札幌
	fmt.Printf("\nFetching pain status for area code: %s (setting weather point: %s)\n", areaCode, setPoint)
	painStatus, err := client.GetPainStatus(areaCode, &setPoint)
	if err != nil {
		log.Fatalf("Error getting pain status: %v", err)
	}
	fmt.Printf("Pain Status for %s:\n", painStatus.PainnoterateStatus.AreaName)
	fmt.Println("------------------------------------")
	fmt.Printf("  Comment: %s\n", painStatus.PainnoterateStatus.Comment)
	fmt.Printf("  Level: %d\n", painStatus.PainnoterateStatus.Level)
	fmt.Println("------------------------------------")

	// 例: Otenki ASPから東京の情報を取得
	otenkiCityCode := "13101" // 東京
	fmt.Printf("\nFetching Otenki ASP data for city code: %s\n", otenkiCityCode)
	otenkiData, err := client.GetOtenkiASP(otenkiCityCode)
	if err != nil {
		log.Fatalf("Error getting Otenki ASP data: %v", err)
	}
	fmt.Printf("Otenki ASP Data (Status: %s, DateTime: %s):\n", otenkiData.Status, otenkiData.DateTime)
	fmt.Println("------------------------------------")
	if len(otenkiData.Elements) > 0 {
		fmt.Printf("Found %d elements.\n", len(otenkiData.Elements))
		// Example: Print today's temperature forecast if available
		for _, elem := range otenkiData.Elements {
			if elem.ContentID == "hight_temp" || elem.ContentID == "low_temp" {
				fmt.Printf("  Element: %s (%s)\n", elem.Title, elem.ContentID)
				today := time.Now().Truncate(24 * time.Hour)
				foundToday := false
				for date, value := range elem.Records {
					if date.Truncate(24 * time.Hour).Equal(today) {
						fmt.Printf("    Today (%s): %v\n", date.Format("2006-01-02"), value)
						foundToday = true
						break // Assuming one value per day for temp
					}
				}
				if !foundToday {
					fmt.Printf("    No data for today.\n")
				}
			}
		}
	} else {
		fmt.Println("  No elements found.")
	}
	fmt.Println("------------------------------------")


	fmt.Println("\nExample usage complete.")
	fmt.Println("To run this example: go run examples/main.go")
}

// Helper function to get environment variable or default
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
