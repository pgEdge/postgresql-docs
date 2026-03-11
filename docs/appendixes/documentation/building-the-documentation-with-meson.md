<a id="docguide-build-meson"></a>

## Building the Documentation with Meson


 Two options are provided for building the documentation using Meson. Change to the `build` directory before running one of these commands, or add `-C build` to the command.


 To build just the HTML version of the documentation:

```

build$ ninja docs
```
 To build all forms of the documentation:

```

build$ ninja alldocs
```
 The output appears in the subdirectory `build/doc/src/sgml`.
