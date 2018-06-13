# MiataChallenge Bridge

A [PortableCloud OS](https://nucleos.com) application (works standalone as a Go
program too) designed to ingest readouts from a Motorola FX9500 reader, save
them to a local SQLite3 database and upload them to an endpoint specified in a
flag using a very primitive custom protocol.

The modem's IP is hardcoded to `192.168.1.151` due to our legacy equipment's
constraints.

In order to use the software, connect to a PCOS device, update flags in the
`command` field in the `portablecloud.[arch].json` manifest and run the following
set of devtools commands command:

```
devtools build -f portablecloud.[arch].json .
devtools device install package.tar
```

After a short while, you should be able to access the UI at
`http://miatachallenge-bridge.my.pcloud.host`. From there, you will be able
to monitor every component of the service.
