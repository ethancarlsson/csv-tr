package cmd

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func getColumnMap(args []string) ([][2]int, error) {

	colMap := make([][2]int, 0, len(args))

	errors := make([]string, 0)
	for _, s := range args {
		if strings.Contains(s, "=") {
			split := strings.SplitN(s, "=", 2)

			if len(split) != 2 {
				errors = append(errors, fmt.Sprintf("Invalid format %s. Contained no = sign.", s))
				continue
			}

			from, err := strconv.ParseInt(split[0], 10, 64)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid format %s", s))
				continue
			}

			to, err := strconv.ParseInt(split[1], 10, 64)

			if err != nil {
				errors = append(errors, fmt.Sprintf("Invalid format %s", s))
			}

			colMap = append(colMap, [2]int{int(from), int(to)})
		}
	}

	if len(errors) > 0 {
		return colMap, fmt.Errorf("All arguments must have the format {from}={to_column}, the right side must be an integer.\n%s", strings.Join(errors, "\n"))
	}

	return colMap, nil
}

func buildColumnFilterMap(filterOut []string) (map[int][]string, error) {
	filterMap := make(map[int][]string, 0)
	errors := make([]string, 0)

	for _, s := range filterOut {
		split := strings.SplitN(s, "=", 2)

		if len(split) != 2 {
			errors = append(errors, fmt.Sprintf("Invalid format %s. Contained no = sign", s))
			continue
		}

		col, err := strconv.ParseInt(split[0], 10, 64)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Invalid format %s, left side must be an integer", s))
			continue
		}

		filterMap[int(col)] = append(filterMap[int(col)], split[1])

	}

	if len(errors) > 0 {
		return filterMap, fmt.Errorf("All arguments to --filter-out or -f must have the format {column}={string}.\n%s", strings.Join(errors, "\n"))
	}

	return filterMap, nil
}

var filterOut []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "csv-tfr",
	Short: "Take a csv from stdin and [t]ranform, [f]ilter it then [r]emember for the next run",
	Long: `This tool takes a csv as input and outputs a new transformed csv.
	The tool allows for interactive filtering and can remember the results
	of previous interactive sessions so the next run can be partially automated
	`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		// we process the input reader, wherever to be his origin
		sep := cmd.Flag("seperator")
		filterInteractiveFlag := cmd.Flag("filter-interactive").Value.String()
		interactiveFilterCol, err := strconv.ParseInt(filterInteractiveFlag, 10, 64)

		shouldFilterInteractively := filterInteractiveFlag != "-1"

		if interactiveFilterCol < 0 {
			return fmt.Errorf("filter-interactive needs to use a column within the csv, column %d does not exist", interactiveFilterCol)
		}

		if err != nil {
			return fmt.Errorf("filter-interactive must accept an integer, %s given", filterInteractiveFlag)
		}

		filterOutMap, err := buildColumnFilterMap(filterOut)

		if err != nil {
			return err
		}
		reader := bufio.NewReader(cmd.InOrStdin())

		newCsv := make([]string, len(args))

		colMap, err := getColumnMap(args)

		if err != nil {
			return err
		}

		// Set the size here so if we find we need to resize we can
		// mutate this value instead of resizing for every row
		newLineSize := len(args)

		// stderr := cmd.OutOrStderr()

		lines := make([]string, 0, 1000)
		var v string
		for err == nil {

			v, err = reader.ReadString('\n')

			lines = append(lines, v)
		}

		// Open TTY device file
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err != nil {
			return fmt.Errorf("Couldn't open TTY for user input: %s", err)
		}
		buf := make([]byte, 1, 1024)

		stderr := cmd.OutOrStderr()

		for _, l := range lines {
			shouldAppend := true
			line := strings.Split(string(l), sep.Value.String())

			for from, list := range filterOutMap {
				if from >= len(line) || from < 0 {
					continue
				}

				if slices.Contains[[]string, string](list, line[from]) {
					shouldAppend = false

				}
			}

			// This block only runs when running in interactive
			if shouldFilterInteractively {
				stderr.Write([]byte(l))
				stderr.Write([]byte("y/n"))

				_, err := tty.Read(buf)

				if err != nil {
					return fmt.Errorf("Couldn't read TTY for user input: %s", err)
				}

				if string(buf[0]) != "y" {
					stderr.Write([]byte("✕\n"))

					filterOutMap[int(interactiveFilterCol)] = append(filterOutMap[int(interactiveFilterCol)], line[interactiveFilterCol])

					continue

				}
				stderr.Write([]byte("✓\n"))

				buf = make([]byte, 1024)
			}

			newLine := make([]string, newLineSize, 100)

			if !shouldAppend {
				continue
			}

			for _, fromTo := range colMap {
				from := fromTo[0]
				to := fromTo[1]

				if to > len(newCsv) {
					newLine = newLine[:to+1]
					newLineSize = to + 1
				}

				if from >= len(line) || from < 0 {
					continue
				}

				newLine[to] = line[from]
			}

			if shouldAppend {
				newCsv = append(newCsv, strings.Join(newLine, sep.Value.String()))
			}
		}

		defer tty.Close()

		writer := cmd.OutOrStdout()
		_, err = writer.Write([]byte(strings.Join(newCsv, "\n")))
		return err
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.csv-transform-remember.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("seperator", "s", ",", "Provide a different seperator for other formats like psv or tsv.")

	strSlice := make([]string, 0)
	rootCmd.Flags().StringSliceVarP(&filterOut, "filter-out", "f", strSlice, "-f {col_number}={string}")

	rootCmd.Flags().IntP(
		"filter-interactive",
		"i",
		-1,
		"-i {number}. Where {number} is the column in the original csv that should be checked for filtering. Negative numbers or numbers greater than the amount of columns in the original csv will be ignored.",
	)
}
