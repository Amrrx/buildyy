// pkg/reporting/reporting.go
package reporting

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"buildy/pkg/config"
)

type BuildReport struct {
	Timestamp   time.Time
	SubProjects []SubProjectReport
}

type SubProjectReport struct {
	Name    string
	Version string
	Status  string
	Error   string
}

func GenerateBuildReport(cfg *config.Config, buildResults map[string]error) (*BuildReport, error) {
	report := &BuildReport{
		Timestamp:   time.Now(),
		SubProjects: make([]SubProjectReport, len(cfg.SubProjects)),
	}

	for i, subProject := range cfg.SubProjects {
		report.SubProjects[i] = SubProjectReport{
			Name:    subProject.Name,
			Version: subProject.Version,
			Status:  "Success",
			Error:   "",
		}

		if err, ok := buildResults[subProject.Name]; ok && err != nil {
			report.SubProjects[i].Status = "Failure"
			report.SubProjects[i].Error = err.Error()
		}
	}

	return report, nil
}

func SaveBuildReport(report *BuildReport, outputDir string) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	timestamp := report.Timestamp.Format("20060102150405")
	filename := fmt.Sprintf("build_report_%s.txt", timestamp)
	filePath := filepath.Join(outputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating report file: %v", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "Build Report - %s\n\n", report.Timestamp.Format(time.RFC3339))
	for _, subProject := range report.SubProjects {
		fmt.Fprintf(file, "Subproject: %s\n", subProject.Name)
		fmt.Fprintf(file, "Version: %s\n", subProject.Version)
		fmt.Fprintf(file, "Status: %s\n", subProject.Status)
		if subProject.Error != "" {
			fmt.Fprintf(file, "Error: %s\n", subProject.Error)
		}
		fmt.Fprintln(file)
	}

	return nil
}
