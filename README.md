# redgiant

Go and REST API for [Sungrow] inverters. Because when the sun starts to grow, the next stage is a [red giant](https://en.wikipedia.org/wiki/Red_giant).

# Are there any prerequisites?

Your [Sungrow] inverter needs to be connected to your network, either through ethernet or wifi, and accessible from the host you want to run redgiant on. To be able to connect to the inverter, you need to find the hostname of the inverter. In most cases this is a local IP address starting with `192.168.XXX.YYY`. This step very much depends on your network so there is no standard way to it. You can for example look at all IP addresses listed in your router and paste them in your browser. You know you found the [Sungrow] inverter if

1. your browser displays a security warning about a self-signed certificate, and
2. you see an orange-branded login dialog after you continue in your browser by accepting the security risk.

The hostname or IP address that you found is referenced by `$SUNGROW_HOST` below.

## How can I login

There are two accounts available to login into your [Sungrow] inverter:

- regular account with username `user` with the default password `pw1111`, and
- admin account with username `admin` with the default password `pw8888`.

redgiant does *not* need elevated permissions so you can user either account.

# How do I use it?

## Go API

## Standalone binary

## Docker

```shell
docker run -e SUNGROW_HOST=$SUNGROW_HOST -p 8000:80 ghcr.io/pmeier/redgiant:latest
```

[Sungrow]: https://en.sungrowpower.com/
