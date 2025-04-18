# CDS lock remover

Clear .cdslock in cellviews. Sometimes when you accidentally crash Cadence Virtuoso session, cellviews that you are working on might be locked which prevent you to edit. To unlock it, you need to manually remove the .cdslock files inside the cellview folder which can be tiresome. Hence, we create a script that helps you remove .cdslock files.

## Installation

If you installed Go (and make), you can build an executable binary file with command

```bash
make build-linux
```

You will get an executable binary file cdslock-remove, which can be coppied over to your Linux server and be executed from there.

Alternatively, you can download the binary file from:

## TODO

- handle the case when user add -folder flag
- showing default of /home/$USER or $CDR_PRJ
