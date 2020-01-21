# codechecker #

A simple tool to find regular expression match in codebase.

## Compilation ##
```
go build codechecker.go
```

## get started ##

Copy the file as below:
```
cp ./synconfig.json <path of your project>/.synconfig.json
cp ./rules.json <path of your project>/.rules.json
```

The file .synconfig.json contain 1 regex : comment and 2 array of regex: filter and exclude.  
The regex "comment" is used to exclude comment line from syntax check.
If empty do not exclude comment with comment.
The array "filter" is used to include only some file who contain one of the entry.  
If empty, take all file.  
The array "exclude" is used to exclude some file who contain one of the entry.  
If empty, take all file.  

Run:
```
./codechecker <path of your project>
```

## Example config ##

.rules.json
```
[
  {
    "rule": "\t+",
    "message": "Tabulation detected"
  },
  {
    "rule": "\\( ",
    "message": "Useless space after parenthese",
    "comment": false
  }
]
```

.synconfig.json
```
{
    "comment": " */\\*([^*]|[\\r\\n]|(\\*+([^*/]|[\\r\\n])))*\\*+/",
    "filter": [
        ".\\.c$",
        ".\\.h$"
    ],
    "exclude": [
        "src_ut/"
    ]
}
```

