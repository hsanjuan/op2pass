package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type entry struct {
	Details  details  `json:"details"`
	Overview overview `json:"overview"`
}

type details struct {
	Fields     []field   `json:"fields"`
	NotesPlain string    `json:"notesPlain"`
	Sections   []section `json:"sections"`
}

type field struct {
	Designation string `json:"designation"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Value       string `json:"value"`
}

type overview struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type section struct {
	Fields []sectionField `json:"fields"`
	Name   string         `json:"name"`
	Title  string         `json:"title"`
}

type sectionField struct {
	K string `json:"k"` // type (string/phone etc)
	N string `json:"n"` // name
	T string `json:"t"` // tag - seems to have a human string
	V *value `json:"v"` // value
}

type value string

func (v *value) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		var i int
		err = json.Unmarshal(b, &i)
		if err != nil {
			*v = value(b)
		}
		*v = value(strconv.Itoa(i))
	} else {
		*v = value(str)
	}
	return nil
}

func main() {
	var e entry

	if len(os.Args) < 2 {
		panic("need argument")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&e)
	if err != nil {
		panic(err)
	}

	titleOpts := make(map[string]struct{})
	titleOpts[e.Overview.Title] = struct{}{}
	titleURL, err := url.Parse(e.Overview.Title)
	if err == nil {
		if titleURL.Host != "" {
			titleOpts[titleURL.Host] = struct{}{}
		} else if titleURL.Path != "" {
			titleOpts[titleURL.Path] = struct{}{}
		}
	}

	titleURL, err = url.Parse(e.Overview.URL)
	if err == nil {
		if titleURL.Host != "" {
			titleOpts[titleURL.Host] = struct{}{}
		} else if titleURL.Path != "" {
			titleOpts[titleURL.Path] = struct{}{}
		}
	}

	for k := range titleOpts {
		new := strings.TrimPrefix(k, "www.")
		if new != k {
			titleOpts[new] = struct{}{}
			delete(titleOpts, k)
		}
	}

	var title string
	var ok bool
	if len(titleOpts) == 0 {
		panic("no title")
	}
	if len(titleOpts) > 1 {
		m := make(map[string]string)
		i := 1
		fmt.Println("select title:")
		for k := range titleOpts {
			m[strconv.Itoa(i)] = k
			fmt.Println(i, " -   ", k)
			i++
		}
		var input string
		fmt.Printf("Choose: ")
		_, err = fmt.Scanln(&input)
		if err != nil {
			panic(err)
		}
		title, ok = m[input]
		if !ok {
			panic("not ok")
		}
	} else {
		for k := range titleOpts {
			title = k
		}
	}

	fileName := fmt.Sprintf("  %s/%s", title, findUsername(e))
	var b strings.Builder

	pass := findPassword(e)
	if len(pass) > 0 {
		b.WriteString(pass + "\n")
	}

	for k, v := range findOtherFields(e) {
		b.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	cmd := exec.Command("pass", "insert", fileName, "-m")
	cmd.Stdin = strings.NewReader(b.String())
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(out.String())
}

func findUsername(e entry) string {
	fields := e.Details.Fields
	for _, f := range fields {
		if f.Designation == "username" {
			return f.Value
		}
	}
	return ""
}

func findPassword(e entry) string {
	fields := e.Details.Fields
	for _, f := range fields {
		if f.Designation == "password" {
			return f.Value
		}
	}
	return ""
}

func findOtherFields(e entry) map[string]string {
	fields := make(map[string]string)
	for _, f := range e.Details.Fields {
		if f.Designation == "username" || f.Designation == "password" {
			continue
		}
		if f.Value == "" {
			continue
		}

		fields[f.Name] = f.Value
	}

	for _, s := range e.Details.Sections {
		for _, f := range s.Fields {
			if f.V != nil && string(*f.V) != "" {
				fields[f.T] = string(*f.V)
			}
		}
	}
	return fields
}
