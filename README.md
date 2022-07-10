# dockerchk
Checks DockerHub if any of the running images require an update. If yes, outputs the IDs of all containers that require updates.

Please note: if this program is started as user root, privileges will be dropped to "nobody" after fetching the list of running containers (use `--nobody-username` to specify another user).

### Build
This program makes use of the syscall.Setuid and syscall.Setgid interfaces to drop privileges so it has to be built with Go 1.16+. In previous versions, this function is not implemented and returns an error.

```
go build .
```

### Usage
``` bash
./dockerchk # check all running containers for available updates
./dockerchk -i onlyoffice/documentserver # check only if running onlyoffice/documentserver containers require an update
```

### Exit codes

* 0: no updates required
* 1: updates could not be checked, an error occured
* 2: at least one update required