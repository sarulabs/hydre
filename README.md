# Hydre

Hydre allows you to run several commands in a docker container.

Usually you only run one process in a docker container. It is the best practice. But sometimes you may still want to run several processes in one container because :

- the processes are tightly coupled and/or it is really complicated to separate them.
- it does not bring anything more to split these processes in different containers in terms of scalability.

#### Use case :

- php and nginx
- a small mail server (postfix + dovecot + spamassassin + amavis + clamav)

A php and nginx example is available [in the example directory](https://github.com/sarulabs/hydre/example/php-nginx).


## Behavior

Hydre starts several daemons. It will run as long as every daemon is working properly. When one daemon dies Hydre stops. You have to restart the container to restart the daemons.


## Installation

To run Hydre you need two files :
- the Hydre binary
- a configuration file

It is easy to include it in a Dockerfile :

```
COPY hydre.yml /home/hydre.yml
ADD http://github.com/sarulabs/hydre/bin/hydre-2.0.0 /home/hydre

CMD ["/home/hydre", "-c", "/home/hydre.yml"]
```


## Configuration file

The configuration file is in `yaml` format.

```yml
timeout: 3
daemons:
    myAwesomeApp:
        command: "my_awesome_app start"
    coolDaemon:
        command: "/ect/init.d/cool_daemon start"
        stopCommand: "/ect/init.d/cool_daemon stop"
        pidFile: "/var/run/cool_daemon/pid"
        logFiles: ["/var/log/cool_daemon.access", "/var/log/cool_daemon.error"]
```

- ***timeout*** : time in second that daemons have to stop gracefully
- ***daemons*** : the definition of the commands you want to execute
    - ***command*** : the unix command that start the program
    - ***stopCommand*** : a command to stop the program gracefully (optional)
    - ***pidFile*** : the path to the program pid file (optional)
    - ***logFiles*** : an array of paths to log files (optional)

### Foreground process

If you want to execute a program that runs in the foreground, you only need to specify the `command` parameter. If the logs are not written directly on the standard output, it is possible to stream them on the standard output of the container with the `logFiles` parameter.

### Background process

If you want to execute a program that runs in the background (init.d scripts for example) you need to specify a `pidFile`. You can also set the `stopCommand` parameter to stop the program gracefully and avoid keeping files from shared volumes in an undesired state.
