package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// downloadFile downloads a file from a URL to a local path
func downloadFile(url, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// extractRepoSection extracts a specific repository section from the content
func extractRepoSection(content, repoID string) string {
	pattern := regexp.MustCompile(`(?m)^\[` + regexp.QuoteMeta(repoID) + `\](.*?)(?:^\[|\z)`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[0]
	}
	return ""
}

// extractAllRepoSections extracts all repository sections from the content
func extractAllRepoSections(content string) map[string]string {
	result := make(map[string]string)

	// Find all repository section headers
	repoHeaderPattern := regexp.MustCompile(`(?m)^\[(.*?)\]`)
	repoHeaders := repoHeaderPattern.FindAllStringSubmatch(content, -1)

	for i, header := range repoHeaders {
		if len(header) < 2 {
			continue
		}

		repoID := header[1]
		var sectionEnd int

		// If this is not the last section, find the next section start
		if i < len(repoHeaders)-1 {
			nextHeaderIndex := strings.Index(content, "["+repoHeaders[i+1][1]+"]")
			sectionEnd = nextHeaderIndex
		} else {
			sectionEnd = len(content)
		}

		// Find the start of this section
		sectionStart := strings.Index(content, "["+repoID+"]")
		if sectionStart >= 0 && sectionEnd > sectionStart {
			result[repoID] = content[sectionStart:sectionEnd]
		}
	}

	return result
}
