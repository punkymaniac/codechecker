package main

import (
    "fmt"
    "io/ioutil"
    "regexp"
    "strings"
    "path/filepath"
    "os"
    "encoding/json"
)

type Config struct {
    // Regex string to catch comment
    Comment string

    // Array of regex to filter the path or files
    Filter []string

    // Array of regex to exclude some path or files
    Exclude []string

    // The regex string compiled
    RegexComment *regexp.Regexp
}

type Rule struct {
    // Regex string
    // Rule of bad syntax to check
    Rule string `json:"rule"`

    // Message printed when the rule is not respected
    Message string `json:"message"`

    // Include comment to search the regex match
    // Optional, default value: true
    Comment bool `json:"comment"`

    // Compiled regex
    Regex *regexp.Regexp `json:regex`
}

// UnmarshalJSON method of the Rule object
// Allow us to deal with specific or optional value
func (r *Rule) UnmarshalJSON(data []byte) error {
    var m map[string]interface{}
    err := json.Unmarshal(data, &m)
    if err != nil {
        return err
    }

    // Check if the optional value is given
    _, ok := m["comment"]
    if !ok {
        // If not given, init it with the default value
        m["comment"] = true
    }

    // Fill the object with the value
    r.Rule = m["rule"].(string)
    r.Message = m["message"].(string)
    r.Comment = m["comment"].(bool)

    return nil
}

// Get files from given directory, filter by a list of regex
func GetFiles(
    path string, // Directory to list file
    filter []string, // A list of regex to filter the return filename
) []string {
    fail := false
    files := []string{}
    compFilter := []*regexp.Regexp{}

    // Compile all filter regex
    for _, f := range filter {
        comp, err := regexp.Compile(f)
        if err != nil {
            fail = true
            fmt.Println(err)
        } else {
            compFilter = append(compFilter, comp)
        }
    }
    if fail == true {
        return files
    }

    // Get all filename in the given path
    // Use function to filter the filename
    err := filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        if len(compFilter) != 0 {
            // Loop over filter to see if the filename match an regex
            for _, reg := range compFilter {
                match := reg.FindAllString(filename, -1)
                // Keep the filename if match it
                if len(match) > 0 {
                    files = append(files, filename)
                }
            }
        } else {
            // If no filter keep all files
            files = append(files, filename)
        }
        return nil
    })
    if err != nil {
        fmt.Println(err)
    }
    return files
}

// Remove filename who match a least one regex
func ExcludeFiles(
    exFiles []string, // The file to apply the exclude filter
    exclude []string, // A list of regex use to exclude some file
) []string {
    fail := false
    files := []string{}
    compExclude := []*regexp.Regexp{}

    // Compile all exclude regex
    for _, f := range exclude {
        comp, err := regexp.Compile(f)
        if err != nil {
            fail = true
            fmt.Println(err)
        } else {
            compExclude = append(compExclude, comp)
        }
    }
    if fail == true {
        return files
    }

    // Loop over filename
    for _, filename := range exFiles {
        keep := true
        for _, reg := range compExclude {
            // If filename match a least one exclude regex don't keep it
            match := reg.FindAllString(filename, -1)
            if len(match) > 0 {
                keep = false
            }
        }
        if keep == true {
            files = append(files, filename)
        }
    }
    return files
}

func main() {

    // Check if path is given
    if len(os.Args) != 2 {
        fmt.Printf("%s <path>\n", os.Args[0])
        return
    }
    path := os.Args[1]
    if path[len(path)-1:] != "/" {
        path += "/"
    }

    config := Config{}

    // Get config from the config file
    conf, err := ioutil.ReadFile(path + ".synconfig.json")
    if err != nil {
        fmt.Println(err)
    }
    err = json.Unmarshal(conf, &config)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Compile the regex to found comment
    config.RegexComment, err = regexp.Compile(config.Comment)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Get rules from rules file
    fr, err := ioutil.ReadFile(path + ".rules.json")
    if err != nil {
        fmt.Println(err)
    }
    rawRules := []Rule{}
    err = json.Unmarshal(fr, &rawRules)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Compile all regex rule
    fail := false
    rules := []Rule{}
    for _, r := range rawRules {
        reg, err := regexp.Compile(r.Rule)
        if err != nil {
            fail = true
            fmt.Println(err)
        } else {
            r.Regex = reg
            rules = append(rules, r)
        }
    }
    // If a least one error of regex compilation, close the program
    if fail == true {
        return
    }

    // Get files to check syntax, who match the config.Filter regex
    files := GetFiles(path, config.Filter)

    // Exclude some file or path, who match the config.Exclude regex
    files = ExcludeFiles(files, config.Exclude)

    // Loop over each files
    for _, file := range files {
        // Remove absolute path from filename
        filename := strings.Replace(file, path, "", -1)
        if filename[0] == '/' {
            filename = filename[1:]
        }

        raw, err := ioutil.ReadFile(file)
        if err != nil {
            fmt.Printf("%s: ERROR\n %v\n", filename, err)
        } else {
            content := string(raw)
            line := strings.SplitAfter(content, "\n")

            //
            // Check rules
            //
            if len(rules) > 0 {

                // Get index start, end of each comment
                // So we know the section who are commented
                commentIndex := config.RegexComment.FindAllStringIndex(content, -1)
                nbrCommentIdx := len(commentIndex)
                for _, r := range rules {
                    // Get index of all rules match in the file
                    matchIdx := r.Regex.FindAllStringIndex(content, -1)
                    // Get all rules match in the file
                    match := r.Regex.FindAllString(content, -1)

                    for nbm, _ := range match {
                        comment := false
                        // If comment are exclude to check the rule
                        if r.Comment == false {
                            // Check if the match is in a comment section
                            for j := 0; j < nbrCommentIdx; j++ {
                                if matchIdx[nbm][0] >= commentIndex[j][0] && matchIdx[nbm][0] <= commentIndex[j][1] {
                                    comment = true
                                    break
                                }
                            }
                        }
                        // If not in comment section
                        if comment == false {
                            nbl := 0
                            // Get number of current line
                            for i := 0; i < matchIdx[nbm][0]; i++ {
                                if content[i] == '\n' {
                                    nbl++;
                                }
                            }

                            // Compute index of the issue on the line
                            idx := 0
                            for idx = matchIdx[nbm][0]; idx != 0 && content[idx] != '\n'; idx-- {}

                            idx = matchIdx[nbm][0] - idx

                            // Print issue
                            fmt.Printf("%s:%d: %s\n%s", filename, nbl + 1, r.Message, line[nbl])
                            for i := 0; i < idx; i++ {
                                fmt.Printf(" ")
                            }
                            fmt.Printf("\033[32;01m^\033[0m\n")
                        }
                    }
                }
            }
        }
    }
}

