## mdai update

update a configuration

### Synopsis

update a configuration file or edit a configuration in an editor

```
mdai update [-f FILE] [--config CONFIG-TYPE] [--phase PHASE] [--block BLOCK]
```

### Examples

```
	mdai update -f /path/to/mdai-operator.yaml  # update mdai-operator configuration from file
	mdai update --config=otel                   # edit otel collector configuration in $EDITOR
	mdai update --config=otel --phase=logs      # jump to logs block
	mdai update --config=otel --block=receivers # jump to receivers block
```

### Options

```
      --block string    block to jump to [receivers, processors, exporters]
  -c, --config string   config type to update
  -f, --file string     file to update
  -h, --help            help for update
      --phase string    phase to jump to [metrics, logs, traces]
```

### SEE ALSO

* [mdai](mdai.md)	 - MyDecisive.ai CLI

###### Auto generated by spf13/cobra on 11-Jul-2024