# Phoenix

Phoenix is an init system for docker.

Usually you only run one process in a docker container. It is the best practice. But sometimes you may still want to run multiple processes in one container because :
- it does not bring you anything more to split these processes in different containers.
- the processes are tightly coupled and it is really complicated to separate them.

To achieve that you need to execute an init system in docker. It is this init system that will be responsible for the different processes you want to run.


## Installation

To run phoenix you need two files.
- the phoenix binary
- a configuration file

You can get phoenix with `go get github.com/sarulabs/phoenix` or download it from [here](https://www.sarulabs.com/downloads/phoenix-1.0.1).

Then it is easy to launch it :

```sh
phoenix run conf.yml
```

In a Dockerfile it may look like this :

```
RUN apt-get install -y wget
RUN wget https://www.sarulabs.com/downloads/phoenix-1.0.1
RUN mv phoenix-1.0.1 /home/phoenix
RUN chmod +x /home/phoenix
ADD conf.yml /home/conf.yml

CMD ["/home/phoenix", "run", "/home/conf.yml"]
```

## Start and stop

When you launch phoenix, it starts all the daemons (processes) that it handles. But you can decide to stop or restart a daemon :

```sh
phoenix stop daemon_name conf.yml
phoenix start daemon_name conf.yml
```

## Monitoring

Phoenix monitors the daemons to ensure they are running if they should. If they crash phoenix will automatically restart them.


## Configuration

The configuration file is in `.yml` format.

Let's say you want to run a mail server in docker. You need postfix and dovecot. The configuration file may be :

```yml
interval: 10
port: 9988
daemons:
    dovecot:
        pidfile: /var/run/dovecot/master.pid
        start: /usr/sbin/dovecot
        stop: /usr/bin/doveadm stop
    postfix:
        pidfile: /var/spool/postfix/pid/master.pid
        start: /usr/sbin/postfix start
        stop: /usr/sbin/postfix stop
```

- ***interval*** : phoenix will check the status of the services
every `interval` seconds.
- ***port*** : phoenix will use that port internally. Choose a port you do not use.
- ***daemons***: the definition of the daemons you want to run.
  - ***daemon_name***: the name of the daemon that will be used with phoenix start/stop
    - ***pidfile***: the pid file of your daemon
    - ***start***: a command to start your daemon
    - ***stop***: a command to stop your daemon

Every daemon needs to have a pid file that will be used to determine if it is running or not. If the program you want to run doesn't have a pid file or a command to stop it, you probably want to create a service by adapting the template `/etc/init.d/skeleton` to your needs.
