package main

import (
	"encoding/xml"
	"os"
	"strings"
)

// Root is the top-level XML element.
type Root struct {
	Items []Item `xml:"item"`
}

// Item represents one <item obj="..."> element — a Splunk object.
type Item struct {
	Obj     string `xml:"obj,attr"`
	Synonym string `xml:"synonym,attr"` // non-empty = alias item, no children
	Cmds    []Cmd  `xml:"cmd"`
	Common  Common `xml:"common"`
}

// Common holds metadata shared across all verbs for an object.
type Common struct {
	URI     string  `xml:"uri"`
	ArgsMap ArgsMap `xml:"argsmap"`
}

// ArgsMap maps CLI flag names to EAI (REST) parameter names.
type ArgsMap struct {
	Args []ArgMapping `xml:"arg"`
}

// ArgMapping is one CLI name → EAI name entry.
type ArgMapping struct {
	CliName string `xml:"cliname,attr"`
	EaiName string `xml:"eainame,attr"`
}

// Cmd represents one <cmd name="..."> element — a verb on an object.
type Cmd struct {
	Name       string    `xml:"name,attr"`
	Synonym    string    `xml:"synonym,attr"` // non-empty = verb alias
	URI        string    `xml:"uri"`
	ImpliedArg string    `xml:"implied_arg_name"`
	EaiID      string    `xml:"eai_id"`
	OfflineOK  *struct{} `xml:"offline_ok"`
	Help       CmdHelp   `xml:"help"`
}

// CmdHelp holds the help text fragments for a verb.
type CmdHelp struct {
	Title    string    `xml:"title"`
	Required []HelpArg `xml:"required>arg"`
	Optional []HelpArg `xml:"optional>arg"`
	Examples []string  `xml:"examples>ex"`
}

// HelpArg is a documented argument.
type HelpArg struct {
	Name string `xml:"name,attr"`
	Text string `xml:",chardata"`
}

// ParseXML reads and parses splunkrc_cmds.xml.
func ParseXML(path string) ([]Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var root Root
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	return root.Items, nil
}

// HTTPMethod returns the HTTP method for a given verb name.
func HTTPMethod(verb string) string {
	switch strings.ToLower(verb) {
	case "list", "show", "display", "status":
		return "GET"
	case "remove":
		return "DELETE"
	default:
		return "POST"
	}
}
