# Simple Page Checker

**This project is currently unstable and subject to change**

This is a simple go application to check live HTTP content against simple text checks.

### Usage

Execute the program from the commandline, Either Passing in a definition file as the first argument or providing the definition JSON directly as the first arg.

```bash
# File usage
./spc example-input.json

# Arg usage, Catting file on command line
./spc $(< example-input.json)

# Arg usage
./spc "{<json_content>}"
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