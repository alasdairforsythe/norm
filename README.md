# norm

Basic text normalization.

### Example:

```golang
norm := norm.NewNormalizer("nfd lines collapse trim")
output, err := norm.Normalize(input_slice)
```

#### Options

`nfd` `lowercase` `accents` `quotemarks` `collapse` `trim` `leading-space` `lines`
