package main

import (
	"os"

	"github.com/spf13/cobra"
	"zutool/api" // Update import path
	"zutool/internal/commands"
	"zutool/internal/presenter"
)

// No global jsonFlag anymore

// Removed newTable function

func main() {
	// Instantiate the API client once
	// Using default values for URLs and timeout for now
	// TODO: Consider making these configurable via flags or config file later
	apiClient := api.NewClient("", "", 0)

	// Define a helper function to create the appropriate presenter based on flags
	getPresenter := func(cmd *cobra.Command) presenter.Presenter {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return &presenter.JSONPresenter{Writer: os.Stdout}
		}
		return &presenter.TablePresenter{Writer: os.Stdout}
	}


	rootCmd := &cobra.Command{
		Use:   "zutool",
		Short: "Get info from zutool <https://zutool.jp/>",
		Long:  "A command-line tool to fetch weather and pain forecast information from zutool.jp.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	painStatusCommand := &cobra.Command{
		Use:     "pain_status [area_code]",
		Aliases: []string{"ps"},
		Short:   "Get pain status forecast by prefecture",
		Long:    "Fetches and displays the pain status forecast for the specified prefecture code.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { // Use anonymous function
			pres := getPresenter(cmd) // Get presenter based on flags
			return commands.RunPainStatus(apiClient, pres, cmd, args) // Call command handler
		},
	}
	painStatusCommand.Flags().StringP("set_weather_point", "s", "", "Set weather point code (e.g., '13113') to get area-specific forecast")
	rootCmd.AddCommand(painStatusCommand)

	weatherPointCommand := &cobra.Command{
		Use:     "weather_point [keyword]",
		Aliases: []string{"wp"},
		Short:   "Search for weather points (locations)",
		Long:    "Searches for weather point locations based on the provided keyword (e.g., city name).",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { // Use anonymous function
			pres := getPresenter(cmd) // Get presenter based on flags
			// Note: We pass cmd here so RunWeatherPoint can access the 'kata' flag
			return commands.RunWeatherPoint(apiClient, pres, cmd, args) // Call command handler
		},
	}
	weatherPointCommand.Flags().BoolP("kata", "k", false, "Include katakana name in the output table")
	rootCmd.AddCommand(weatherPointCommand)

	weatherStatusCommand := &cobra.Command{
		Use:     "weather_status [city_code]",
		Aliases: []string{"ws"},
		Short:   "Get detailed weather status by city",
		Long:    "Fetches and displays detailed weather status (temperature, pressure, etc.) for the specified city code.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { // Use anonymous function
			pres := getPresenter(cmd) // Get presenter based on flags
			// Note: We pass cmd here so RunWeatherStatus can access the 'n' flag
			return commands.RunWeatherStatus(apiClient, pres, cmd, args) // Call command handler
		},
	}
	weatherStatusCommand.Flags().IntSliceP("n", "n", []int{0}, "Specify day number(s) to show (-1 to 2)")
	rootCmd.AddCommand(weatherStatusCommand)

	otenkiAspCommand := &cobra.Command{
		Use:     "otenki_asp [city_code]",
		Aliases: []string{"oa"},
		Short:   "Get weather information from Otenki ASP",
		Long:    "Fetches various weather forecast elements (weather, temp, wind, etc.) from the Otenki ASP service for specific major city codes.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error { // Use anonymous function
			pres := getPresenter(cmd) // Get presenter based on flags
			// Note: We pass cmd here so RunOtenkiAsp can access the 'n' flag
			return commands.RunOtenkiAsp(apiClient, pres, cmd, args) // Call command handler
		},
	}
	otenkiAspCommand.Flags().IntSliceP("n", "n", []int{0, 1, 2, 3, 4, 5, 6}, "Specify forecast day number(s) to show (0 to 6)")
	rootCmd.AddCommand(otenkiAspCommand)

	// Define the --json flag persistently for all commands
	rootCmd.PersistentFlags().BoolP("json", "j", false, "Output results in JSON format")

	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error, so just exit
		// fmt.Fprintln(os.Stderr, err) // Removed redundant error printing
		os.Exit(1) // Exit with error code 1 if execution fails
	}
}

// Removed runPainStatus function
// Removed runWeatherPoint function
// Removed renderWeatherStatusTable function
// Removed runWeatherStatus function
// Removed runOtenkiAsp function
// Removed getMapKeys function
// Removed min function
