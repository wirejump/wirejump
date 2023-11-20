package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

type JSONOutput struct {
	Error   bool        `json:"error"`
	Message interface{} `json:"message"`
}

type OutputFormatter interface {
	FormatOutput(*io.Writer) error
}

// Use it for empty fields
const emptyValue = "N/A"

// Max nested struct depth
const maxNestingLevel = 3

// Get struct fields and values as strings, up to maxNestingLevel levels of recursion
func dataToStringArray(data interface{}, nestingLevel int) ([]string, error) {
	var output []string

	if nestingLevel > maxNestingLevel {
		return []string{}, fmt.Errorf("more than %d nesting levels are not supported", maxNestingLevel)
	}

	if data == nil {
		return []string{}, nil
	}

	// Get value
	v := reflect.ValueOf(data)

	// Extract pointer value if needed
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// Get value type
	t := v.Type()

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			tags := t.Field(i).Tag
			title := t.Field(i).Name

			// Use tag hints if possible
			if tags != "" {
				if lookup, ok := tags.Lookup("pretty"); ok && lookup != "" {
					title = lookup
				}
			}

			if f.Kind() == reflect.Struct {
				nested, err := dataToStringArray(f.Interface(), nestingLevel+1)

				if err != nil {
					return []string{}, err
				}

				output = append(output, fmt.Sprintf("%s\n", title))
				output = append(output, nested...)
				// output = append(outp)
			} else {
				// Get raw value
				value := f.Interface()

				// Extract pointer value if needed
				if f.Kind() == reflect.Pointer {
					if f.IsNil() {
						value = nil
					} else {
						value = f.Elem().Interface()
					}
				}

				// Special formatting for each type
				switch f.Kind() {
				case reflect.String:
					if value == "" {
						value = emptyValue
					}
				case reflect.Bool:
					if value == true {
						value = "yes"
					} else {
						value = "no"
					}
				case reflect.Slice:
					value = strings.Join(value.([]string), ", ")
				}

				// Format time fields
				if tags != "" {
					if _, ok := tags.Lookup("timefield"); ok {
						if value != nil {
							value = prettyTime(value.(int64))
						}
					}
				}

				// Format nil pointers
				if value == nil {
					value = emptyValue
				}

				padding := strings.Repeat(" ", nestingLevel*2)
				output = append(output, fmt.Sprintf("%s%s:\t%v\t\n", padding, title, value))
			}
		}
	} else {
		// Just forward data for non-structs
		output = append(output, fmt.Sprintf("%v\n", v.Interface()))
	}

	return output, nil
}

// Print received reply in a table
func PrettyFormatter(dest io.Writer, data interface{}) error {
	writer := new(tabwriter.Writer)
	as_lines, err := dataToStringArray(data, 0)

	if err != nil {
		return err
	}

	writer.Init(dest, 24, 4, 1, ' ', 0)

	for _, line := range as_lines {
		fmt.Fprint(writer, line)
	}

	writer.Flush()

	return nil
}

// Reencode received JSON with added command status
func JSONFormatter(dest io.Writer, has_error bool, message interface{}) error {
	output := JSONOutput{
		Error:   has_error,
		Message: message,
	}

	enc, err := json.Marshal(output)

	if err != nil {
		return err
	}

	as_string := string(enc)

	// In case original reply has failed, make sure to forward the error
	if has_error {
		return errors.New(as_string)
	}

	fmt.Println(as_string)

	return nil
}
