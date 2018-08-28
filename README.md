# patchy

store and recall jack audio port connections


## store

Produce a json patch from the current connection configuration.

```
# output on stdout
patchy store
```

```
# or to file
patchy store some-file
```


## recall

Create connections from a patch. All ports named in the patch must exist for
`patchy` to actually make or break connections.  All connections not named in
the patch are disconnected.

```
# input from stdin
patchy recall
```

```
# or from file
patchy recall some-file
```

```
# wait on ports and recall
patchy -w 1m recall some-file
```


## build with docker

The `dedelala/patchy-builder` image is built from the Dockerfile
in this repo.

```
docker run -it --rm -v "$PWD:/src" -w /src \
  dedelala/patchy-builder:latest go build
```
