package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var Version string

type Span struct {
	OrgId           int                `json:"org_id"`
	TraceId         string             `json:"trace_id"`
	SpanId          string             `json:"span_id"`
	ParentId        string             `json:"parent_id"`
	Start           float64            `json:"start"`
	End             float64            `json:"end"`
	Duration        float32            `json:"duration"`
	Type            string             `json:"type"`
	Service         string             `json:"service"`
	Name            string             `json:"name"`
	Resource        string             `json:"resource"`
	ResourceHash    string             `json:"resource_hash"`
	HostId          int                `json:"host_id"`
	Env             string             `json:"env"`
	HostGroups      []string           `json:"host_groups"`
	Meta            map[string]string  `json:"meta"`
	Metrics         map[string]float64 `json:"metrics"`
	IngestionReason string             `json:"ingestion_reason"`
	ChildrenIds     []string           `json:"children_ids"`
}

type Trace struct {
	RootId string          `json:"root_id"`
	Spans  map[string]Span `json:"spans"`
}

type TraceData struct {
	Trace       Trace   `json:"trace"`
	Orphaned    []Trace `json:"orphaned"`
	IsTruncated bool    `json:"is_truncated"`
}

type Config struct {
	query      string
	tracePath  string
	verbosity  int
	depthLimit int
	help       bool
	version    bool
}

var rVersionFlag = regexp.MustCompile("^--version$")
var rHelpFlag = regexp.MustCompile("^--help$")
var rVerboseFlag = regexp.MustCompile("^-(?P<verbosity>(v{1,2}))$")
var rDepthPattern = regexp.MustCompile("--depth=(?P<depth>([1-9][0-9]*|0))")

var helpText = fmt.Sprintf(`
Doggo (version %s)

🐕🔎 Inspecting big Datadog traces in the CLI

Usage:

doggo <trace_path> <resource_name or partial string> [-vv] [--depth=<depthLimit>]

Given a trace_path leading to tracedata json and a query (either a resource name or a partial resource name), displays the spans and children of the sections of the trace that match.

By default, timing information for each span is shown.

-v adds more resource-related information for each span;
-vv adds even more resource-related information, including tags, for each span;
--depth=<depthLimit> allows limiting how many levels down each match the display should go.
`, Version)

// Prints traces recursively with increasing ident levels.
// More data is printed until there are no more traces available
// i.e. until the ChildrenIds key yields an empty array.
func recursivelyPrintTraces(idToSpanMap map[string]Span, current string, config Config, depth int) {
	if depth > config.depthLimit {
		return
	}

	currentSpan := idToSpanMap[current]
	prefix := strings.Repeat(" ", depth)

	log.Println(fmt.Sprintf("%s\033[35m[%s]\033[0m: %fms", prefix, currentSpan.Name, currentSpan.Duration*1000))

	if config.verbosity > 0 {
		log.Println(fmt.Sprintf("%s \033[36m> %s\033[0m", prefix, currentSpan.Resource))
	}

	if config.verbosity > 1 {
		meta, err := json.MarshalIndent(currentSpan.Meta, "", "  ")

		if err != nil {
			panic(err)
		}

		log.Println(fmt.Sprintf("%s %s", prefix, meta))
	}

	for _, childId := range currentSpan.ChildrenIds {
		recursivelyPrintTraces(idToSpanMap, childId, config, depth+1)
	}
}

// Builds indexes that make searching by span ID easier.
// The indexes returned are a map of Span structs by span ID and
// a map of span IDs (array) by resource name.
//
// This is sufficient to search by resource name easily.
func buildSpanIndexes(fullTrace TraceData) (map[string]Span, map[string][]string) {
	spansById := map[string]Span{}
	spanIdsByResourceName := map[string][]string{}

	tracesToProcess := []Trace{fullTrace.Trace}

	if fullTrace.Orphaned != nil {
		tracesToProcess = append(tracesToProcess, fullTrace.Orphaned...)
	}

	for _, trace := range tracesToProcess {
		for key, value := range trace.Spans {
			resourceName := value.Resource
			spansById[key] = value

			spanList, ok := spanIdsByResourceName[resourceName]

			if !ok {
				spanIdsByResourceName[resourceName] = []string{key}
			} else {
				spanIdsByResourceName[resourceName] = append(spanList, key)
			}
		}
	}

	return spansById, spanIdsByResourceName
}

// Parses a Datadog-provided trace data JSON into a data
// structure fitting TraceData (see type).
func parseTraceJsonFromFile(path string) TraceData {
	data, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	fullTrace := TraceData{}

	err = json.Unmarshal(data, &fullTrace)

	if err != nil {
		panic(err)
	}

	return fullTrace
}

// Parses command-line arguments to extract flags and
// other useful contextual details.
func parseArgs(args []string) Config {
	rootOfInterest := ""

	if len(os.Args) >= 3 {
		rootOfInterest = os.Args[2]
	}

	tracePath := ""

	if len(os.Args) >= 2 {
		tracePath = os.Args[1]
	}

	verbosity := 0
	depth := 9999
	help := false
	version := false

	verbosityMatchIndex := rVerboseFlag.SubexpIndex("verbosity")
	for _, arg := range os.Args {
		verbosity_match := rVerboseFlag.FindAllStringSubmatch(arg, -1)

		if verbosity_match != nil {
			verbosity = len(verbosity_match[0][verbosityMatchIndex])
			continue
		}

		depthMatch := rDepthPattern.FindStringSubmatch(strings.Join(os.Args, " "))

		if depthMatch != nil {
			depthValueIndex := rDepthPattern.SubexpIndex("depth")

			depthValue := depthMatch[depthValueIndex]

			depthLimit, err := strconv.Atoi(depthValue)

			if err != nil {
				log.Println("Couldn't parse depth limit, ignoring")
			}

			depth = depthLimit
			continue
		}

		if rHelpFlag.MatchString(arg) {
			help = true
		}

		if rVersionFlag.MatchString(arg) {
			version = true
		}

	}

	return Config{rootOfInterest, tracePath, verbosity, depth, help, version}
}

func main() {
	config := parseArgs(os.Args)

	if config.help {
		fmt.Println(helpText)
		return
	}

	if config.version {
		fmt.Println(Version)
		return
	}

	fullTrace := parseTraceJsonFromFile(config.tracePath)

	spansById, spanIdsByResourceName := buildSpanIndexes(fullTrace)

	matchedIds := []string{}

	rRootOfInterest := regexp.MustCompile(config.query)

	log.Println("Looking for spans matching resource_name pattern: " + config.query)
	for key, ids := range spanIdsByResourceName {
		if rRootOfInterest.MatchString(key) {
			matchedIds = append(matchedIds, ids...)
		}
	}
	log.Println(fmt.Sprintf("Found %d traces!", len(matchedIds)))

	for position, traceId := range matchedIds {
		log.Println(fmt.Sprintf("### Trace #%d ###", position))
		recursivelyPrintTraces(spansById, traceId, config, 0)
	}
}
