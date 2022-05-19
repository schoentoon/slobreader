# slobreader

[![Gitlab pipeline status](https://gitlab.com/schoentoon/slobreader/badges/master/pipeline.svg)](https://gitlab.com/schoentoon/slobreader)

A commandline [slob](https://github.com/itkach/slob)reader, specifically aimed at the [freedict](https://freedict.org/downloads/index.html#smartphones-and-tablets) files for translation lookup.
You can launch this tool with either a .slob file as an argument.
Or with a [config file](./german.yml), which can also be used to pretty print genders of words, disable autocomplete, ignore certain entries.
Last but not least, you can also launch it without any arguments, in which case it will attempt to use a config file located at `~/.config/slobreader/default.yml`.

Below is an example of how execution of this tool looks like.
Note how with the first word there was no autocomplete available yet, as this was still loading in the background.

[![asciicast](https://asciinema.org/a/QDDbbBYnjVn0mhNk8noAxc2ky.svg)](https://asciinema.org/a/QDDbbBYnjVn0mhNk8noAxc2ky)
