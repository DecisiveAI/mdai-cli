## mdai completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(mdai completion bash)

To load completions for every new session, execute once:

#### Linux:

	mdai completion bash > /etc/bash_completion.d/mdai

#### macOS:

	mdai completion bash > $(brew --prefix)/etc/bash_completion.d/mdai

You will need to start a new shell for this setup to take effect.


```
mdai completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### SEE ALSO

* [mdai completion](mdai_completion.md)	 - Generate the autocompletion script for the specified shell

###### Auto generated by spf13/cobra on 11-Jul-2024
