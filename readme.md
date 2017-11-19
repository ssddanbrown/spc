# Simple Page Checker

**This project is currently unstable and subject to change**

This is a simple go application to check live HTTP content against simple text checks.

### Usage

Execute the program from the command line, Either Passing in a definition file as the first argument or providing the definition JSON directly as the first arg. Alternatively, If no arguments are provided the definition file can be provided via stdin. 

```bash
# File usage
./spc example-input.json

# Arg usage, Catting file on command line
./spc $(< example-input.json)

# Arg usage
./spc "{<json_content>}"

# stdin usage
cat example-input.json | ./spc
```

### Definition File

The definition file is a json formatted file as the below example:

```json
{
    "checks": {
        ".*\\.me": "Dan",
        "danb": ["Dan", "Brown"],
        "//(.*?)\\.com": "welcome to $1"
    },
    "urls": [
        "https://danb.me",
        "https://example.com",
        "https://github.com/ssddanbrown/haste"
    ]
}
```

The `checks` object keys are regex strings that will be checked against the URL. If the regex matches the URL the check content, provided as the value, will be search in the response content.  
The checks can be either a string or an array of strings to check against.

Any regex matches within the url regex can be inserted into a check using `$1` style placeholders.

### Output

The above definition example will output the following:

```shell
Checking 3 urls, 5 checks

https://danb.me
        ✔ [Dan]
        ✔ [Dan]
        ✔ [Brown]
https://github.com/ssddanbrown/haste
        ✔ [Dan]
        ✔ [Brown]
        ✗ [welcome to github]
https://example.com
        ✗ [welcome to example]

5 checks passed, 2 checks failed, 71.43% of tests passed
```

### Docker Container

A lightweight docker container can be found at https://hub.docker.com/r/ssddanbrown/spc/. This is mainly for usage on CI systems such as GitLab CI. The binary can be found at the path `/spc` of the container.

If you are using the container normally via the command line, The easier way to do this is via piping:

```bash
cat example-input.json | docker run -i ssddanbrown/spc:latest
```


