# Stress tool

This repo fork from influxdata/influx-stress.
Change to stress test on influxdb and mysql/postgress, and do some comparision.

## Build Instructions

Building `dbstress` requires the Golang toolchain. If you do not have the Golang toolchain installed
please follow the instructions [golang.org/doc/install](https://golang.org/doc/install)

```sh
go get -v github.com/deltacat/dbstress/...
```

## Top Level Command

try `--help`

## Example Usage

>Before first launch, please rename sample configure file (dbstress.sample.toml) to dbstress.toml, then make some necessary change according to your environment setting.

Runs forever

```bash
dbstress insert
```

Runs forever writing as fast as possible

```bash
dbstress insert -f
```

Runs for 1 minute writing as fast as possible

```bash
dbstress insert -r 1m -f
```

Writing an example series key with 20,000 series

```bash
dbstress insert -s 20000 
```
