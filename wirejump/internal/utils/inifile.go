package utils

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
)

var INIEntity = regexp.MustCompile(
	/* comment */ `^[[:blank:]]*(?P<comment>#.*)|` +
		/* section */ `^[[:blank:]]*\[(?P<section>([A-Za-z]*))\]|` +
		/*  data   */ `^[[:blank:]]*(?P<data>([A-Za-z]*)[[:blank:]]*\=[[:blank:]]*(.*))`)

// INIPair is a key = value pair inside a section. Cannot be duplicated
type INIPair map[string]string

// INIFile is a section = contents map. Sections can be duplicated, thus
// each map value is an array of values. Unique sections will be at zero
// index, and if sections are duplicated, all of them will be accessible.
type INIFile map[string][]INIPair

// Read INI file at given path and return parsed struct
func ReadINI(path string) (INIFile, error) {
	sections := []string{}
	values := []INIPair{}
	parsed := INIFile{}

	file, err := os.Open(path)

	if err != nil {
		return INIFile{}, err

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection *INIPair

	for scanner.Scan() {
		match := INIEntity.FindStringSubmatch(scanner.Text())

		if match == nil {
			continue
		}

		groups := INIEntity.SubexpNames()
		total := len(groups)

		for i, name := range groups {
			matched := match[i]

			if i == 0 {
				continue
			}

			// Got section match
			if name == "section" && matched != "" {
				// Previous section has ended, store all values
				if currentSection != nil {
					values = append(values, *currentSection)
				}

				// Store new section name
				sections = append(sections, matched)

				// Create new section map
				currentSection = &INIPair{}
			} else if name == "data" && matched != "" {
				// Ignore data outside any section
				if currentSection != nil {
					k := match[total-2]
					v := match[total-1]

					// Save k-v pair for currentSection
					(*currentSection)[k] = v
				}
			}
		}
	}

	// Process last section
	if currentSection != nil {
		values = append(values, *currentSection)
	}

	if err := scanner.Err(); err != nil {
		return INIFile{}, err
	}

	for idx, value := range values {
		section := sections[idx]

		parsed[section] = append(parsed[section], value)
	}

	return parsed, nil
}

func WriteINI(path string, content INIFile) error {
	file, err := os.Create(path)
	if err != nil {
		return err

	}
	defer file.Close()

	file.Truncate(0)
	file.Seek(0, 0)

	names := []string{}
	writer := bufio.NewWriter(file)

	// Get section names
	for name := range content {
		names = append(names, name)
	}

	// Sort section names
	sort.Strings(names)

	// Iterate all sections
	for _, name := range names {
		// Write (duplicated) section content
		for _, inside := range content[name] {
			_, err := writer.WriteString(fmt.Sprintf("[%s]\n", name))

			if err != nil {
				return err
			}

			for k, v := range inside {
				_, e := writer.WriteString(fmt.Sprintf("%s = %s\n", k, v))

				if e != nil {
					return e
				}
			}

			// Add newline after section end
			_, err = writer.WriteString("\n")

			if err != nil {
				return err
			}
		}
	}

	return writer.Flush()
}
