package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	zglob "github.com/mattn/go-zglob"
	"github.com/mgechev/revive/formatter"

	"github.com/BurntSushi/toml"
	"github.com/mgechev/revive/lint"
	"github.com/mgechev/revive/rule"
)

var defaultRules = []lint.Rule{
	&rule.VarDeclarationsRule{},
	&rule.PackageCommentsRule{},
	&rule.DotImportsRule{},
	&rule.BlankImportsRule{},
	&rule.ExportedRule{},
	&rule.NamesRule{},
	&rule.ElseRule{},
	&rule.IfReturnRule{},
	&rule.RangeRule{},
	&rule.ErrorfRule{},
	&rule.ErrorsRule{},
	&rule.ErrorStringsRule{},
	&rule.ReceiverNameRule{},
	&rule.IncrementDecrementRule{},
	&rule.ErrorReturnRule{},
	&rule.UnexportedReturnRule{},
	&rule.TimeNamesRule{},
	&rule.ContextKeyTypeRule{},
	&rule.ContextArgumentsRule{},
}

var allRules = append([]lint.Rule{
	&rule.ArgumentsLimitRule{},
	&rule.CyclomaticRule{},
}, defaultRules...)

var allFormatters = []lint.Formatter{
	&formatter.CLIFormatter{},
	&formatter.JSONFormatter{},
}

func getFormatters() map[string]lint.Formatter {
	result := map[string]lint.Formatter{}
	for _, f := range allFormatters {
		result[f.Name()] = f
	}
	return result
}

func getLintingRules(config *lint.Config) []lint.Rule {
	rulesMap := map[string]lint.Rule{}
	for _, r := range allRules {
		rulesMap[r.Name()] = r
	}

	lintingRules := []lint.Rule{}
	for name := range config.Rules {
		rule, ok := rulesMap[name]
		if !ok {
			panic("cannot find rule: " + name)
		}
		lintingRules = append(lintingRules, rule)
	}

	return lintingRules
}

func parseConfig(path string) *lint.Config {
	config := &lint.Config{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic("cannot read the config file")
	}
	_, err = toml.Decode(string(file), config)
	if err != nil {
		panic("cannot parse the config file: " + err.Error())
	}
	return config
}

func normalizeConfig(config *lint.Config) {
	severity := config.Severity
	if severity != "" {
		for k, v := range config.Rules {
			if v.Severity == "" {
				v.Severity = severity
			}
			config.Rules[k] = v
		}
	}
}

func getConfig() *lint.Config {
	config := defaultConfig()
	if configPath != "" {
		config = parseConfig(configPath)
	}
	normalizeConfig(config)
	return config
}

func getFormatter() lint.Formatter {
	formatters := getFormatters()
	formatter := formatters["cli"]
	if formatterName != "" {
		f, ok := formatters[formatterName]
		if !ok {
			panic("unknown formatter " + formatterName)
		}
		formatter = f
	}
	return formatter
}

func defaultConfig() *lint.Config {
	defaultConfig := lint.Config{
		Confidence: 0.0,
		Severity:   lint.SeverityWarning,
		Rules:      map[string]lint.RuleConfig{},
	}
	for _, r := range defaultRules {
		defaultConfig.Rules[r.Name()] = lint.RuleConfig{}
	}
	return &defaultConfig
}

func getFiles() []string {
	globs := flag.Args()
	if len(globs) == 0 {
		panic("files not specified")
	}

	var matches []string
	for _, g := range globs {
		m, err := zglob.Glob(g)
		if err != nil {
			panic(err)
		}
		matches = append(matches, m...)
	}

	if excludeGlobs == "" {
		return matches
	}

	excluded := map[string]bool{}
	excludeGlobSlice := strings.Split(excludeGlobs, " ")
	for _, g := range excludeGlobSlice {
		m, err := zglob.Glob(g)
		if err != nil {
			panic("error while parsing glob from exclude " + err.Error())
		}
		for _, match := range m {
			excluded[match] = true
		}
	}

	var finalMatches []string
	for _, m := range matches {
		if _, ok := excluded[m]; !ok {
			finalMatches = append(finalMatches, m)
		}
	}

	return finalMatches
}

var configPath string
var excludeGlobs string
var formatterName string
var help bool

var originalUsage = flag.Usage

func init() {
	flag.Usage = func() {
		fmt.Println(banner)
		originalUsage()
	}
	const (
		configUsage    = "path to the configuration TOML file"
		excludeUsage   = "glob which specifies files to be excluded"
		formatterUsage = "formatter to be used for the output"
	)
	flag.StringVar(&configPath, "config", "", configUsage)
	flag.StringVar(&excludeGlobs, "exclude", "", excludeUsage)
	flag.StringVar(&formatterName, "formatter", "", formatterUsage)
	flag.Parse()
}
