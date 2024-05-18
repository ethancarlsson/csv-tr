# csv-tr

csv-tr is a CLI application takes a csv as input and outputs a new transformed
csv, with columns reoganised. The tool allows for interactive filtering and can
remember the results of previous interactive sessions so the next run can be
partially automated.

## Installation

### Go
```
go install github.com/ethancarlsson/csv-tr
```

## Usage

This tool only works on streams, so if you have a csv file you need to get it
out from the file somehow.

```sh
% cat examples/groceries.csv | csv-tr
```

The above will work, but will produce an empty result. To get a result from the
tool you need to map the columns from one csv to the other.

So if `examples/groceries.csv` is

```csv
7,oranges
1,loaf of bread
4,potatoes
```

Then

```sh
% cat examples/groceries.csv | csv-tr 1=0 0=1
```

will map the 0th column to the 1st column of the new csv and the 1st column to
the 0th column of the new csv.

```csv
oranges,7
loaf of bread,1
potatoes,4
```

Only mapped columns will be present in the new csv.

```sh
cat examples/groceries.csv | csv-tr 1=0

oranges
loaf of bread
potatoes
```

### Interactive mode

Using interactive mode you can interactively choose what data is included in the
final result.

```sh
% cat examples/groceries.csv | csv-tr 1=0 -i 1
7,#oranges
Include? (y/n) y
✓
1,loaf of bread
Include? (y/n) n # Exclude loaf of bread
✕
4,potatoes
Include? (y/n) y
✓
# Results start below output to stdout only including oranges and potatoes
oranges
potatoes
1=loaf of bread # exclusion filter list is output to stderr so that it can be used in a future round
1=oranges,1=potatoes # inclusion filter list is also output to stderr
```

If you use the list filters in stderr those lines will be skipped in interactive
filtering.

```sh
% cat examples/groceries.csv | csv-tr  0=1 1=0 -i 1 --filter-out "1=loaf of bread,1=oranges"
4,potatoes
Include? (y/n) y
✓
potatoes,4
potatoes1=loaf of bread,1=oranges1=potatoes
```

If there are duplicates it will filter out all future results.

### Files

You can store the filters in files rather than stderr by using the `--store-filters`
flag.

```sh
% cat examples/groceries.csv | csv-tr  0=1 1=0 -i 1 --store-filters .
7,oranges
Include? (y/n) y
✓
1,loaf of bread
Include? (y/n) Include? (y/n) y
✓
4,potatoes
Include? (y/n) n
✕
oranges,7
loaf of bread,1

% ls
examples     fin.txt      fout.txt
```

These files can then be referenced instead of passing in the whole string with a
flag.

```sh
% cat examples/groceries.csv | csv-tr  0=1 1=0 -i 1 --store-filters .
oranges,7
loaf of bread,1
```

## Use cases

This is a very particular program, so what is it actually used for? I developed
the program because I share expenses with my partner, I often go through my bank
account statements adding the transactions that need to be shared and leaving out
thoses that don't need to be included. Often, there are things like the local 
groceries that are always shared or things that are never shared. This tool lets
me skip all those cases repeated transactions.

