package commands

import (
	"fmt"
	"errors"
	"zutool/internal/api"
	"zutool/internal/presenter"

	"github.com/spf13/cobra"
)

// RunWeatherPoint executes the logic for the 'weather_point' command.
func RunWeatherPoint(client *api.Client, pres presenter.Presenter, cmd *cobra.Command, args []string) error {
	// 1. Parse arguments and flags
	keyword := args[0]
	kataFlag, _ := cmd.Flags().GetBool("kata") // Get the 'kata' flag value

	// 2. Call the API client
	res, err := client.GetWeatherPoint(keyword)
	if err != nil {
		// Handle potential API errors, e.g., not found for the keyword
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			// If it's a 404, maybe just inform the user via the presenter?
			// Or return a specific error message. For now, return the API error.
			// Let's try printing a message directly for now, as presenter doesn't have a dedicated method.
			// TODO: Consider adding a generic `PresentError` or `PresentMessage` to the presenter interface.
			fmt.Printf("地域地点の検索に失敗しました: %s\n", err.Error()) // Print error directly
			return nil // Return nil to avoid cobra printing the error again
		}
		// For other errors, return them wrapped
		return fmt.Errorf("地域地点の検索中にエラーが発生しました: %w", err)
	}

	// 3. Pass data to the presenter
	err = pres.PresentWeatherPoint(res, kataFlag, keyword)
	if err != nil {
		return fmt.Errorf("結果の表示に失敗しました: %w", err)
	}

	return nil
}
