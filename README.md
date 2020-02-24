# dockermanifestv2reader


Compile CLI:

```
$ go build cmd/main.go
```

Set credentials:

```
export USER_TOKEN=$USER:$PASSWORD
```

Fetch manifest V2 from docker image:

```
$ ./main registry.redhat.io/rhscl/postgresql-10-rhel7:1-47 2>/dev/null
sha256:de3ab628b403dc5eed986a7f392c34687bddafee7bdfccfd65cecf137ade3dfd
```
