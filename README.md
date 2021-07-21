# dockerchk
Checks DockerHub if any of the running images require an update. If yes, outputs the IDs of all containers that require updates.

Please note: if this program is started as user root, privilege will be dropped to "nobody" after fetching the list of running containers.

This program makes use of the syscall.Setuid and syscall.Setgid interfaces to drop privileges so it has to be built with Go 1.16+. In previous versions, this function is not implemented and returns an error.

### Build:
```
go build .
```

### Usage:
``` bash
./dockerchk # check all running containers for available updates
./dockerchk -i onlyoffice/documentserver # check only if running onlyoffice/documentserver containers require an update
```
