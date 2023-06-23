# subfilter

This subfilter middleware allows you to perform regular expression-based search and replace operations on both the body content and the headers of HTTP responses, with optional support for case-insensitive replacements.

## Configuration

Dynamic:
```yaml
http:
  middlewares:
    mysubfilter:
      plugin:
        subfilter:
          replacements:
            - pattern: "Kiwi"
              replacement: "Cherry"
            - pattern: "apple(.)"
              replacement: "banana?"
              flags: "i"
```


Replace `pattern` with `replacement` (Optional: `flags: "i"` specifies a case-insensitive match)
