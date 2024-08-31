# Config
Simple configuration loader


## Why yet another configuration handler

When I was looking for a configuration management, I did not find somthing that 
I liked, specially I struggled with the following points:

- I was not trivial to make a 12 factor app
- unmarshal the configuration to a struct
- support to load configs from a conf.d pattern (similar linux)
- I wanted to add validation actions like required fields

For that reason i started writing this one.


## Main features
* load your configuration into one struct
* use default configuration
* load configuration from files on disk
* use environment variables to overwrite configuration


## Use

Import the module
```
go get github.com/go-bumbu/config
```

Example taken from example test

```
_,err   = config.Load(
    config.Defaults{Item: defaultCfg},                                   // use default values
    config.CfgFile{Path: "sampledata/example_test/example.config.json"}, // load config from file
    config.EnvVar{Prefix: "ENVPREFIX"},                                  // load a config value from an env
    config.Unmarshal{Item: &cfg},                                        // marshal result into cfg
)
```

Multiple config flags can be passed to the load function, and they will enable different features.


## Struct annotations

Struct fields can be annotated to specify the field name in the configuration, e.g.

```
type customCfg struct {
	U string   `config:"user"`
	Pw string  `config:"password"`
}
```

## Best practices

when mapping your configuration there are some consideration that will make your live easier:

### Maps instead of slices
If you need a list of items it's recommended to use a map instead of a slice.

#### Explanation:
picture this scenario: you define a set of default values as a slice 

```
val := []string{
    "a","b"
}
```

and now as an Env you want to add another one, you would need to both provide the original default + the new value

```
#!/bin/bash
export VAL="a,b,c"
```

now let's see the same scenario as a map
```
val := map[string]string{
    "a": "a",
    "b": "b",
}
```
the corresponding env var would be.
```
#!/bin/bash
export VAL_C="c"
```

this is specially important when you have a map of structs 
