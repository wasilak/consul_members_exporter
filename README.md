# consul_members_exporter [![Build Status](https://travis-ci.com/wasilak/consul_members_exporter.svg?branch=main)](https://travis-ci.com/wasilak/consul_members_exporter) [![Maintainability](https://api.codeclimate.com/v1/badges/de65d3d71ee04587d568/maintainability)](https://codeclimate.com/github/wasilak/consul_members_exporter/maintainability)
Prometheus exporter providing details about known members

This is meant to be supplementary exporter for Consul providing information about known memebers.

Provided metrics:
* `consul_members_details` gauge with constant value `1` providing information as labels, i.e.: 
  ```
  consul_members_details{addr="192.168.50.4",name="consul-3.local",server="true",status="1",statusText="Alive",version="1.9.0"} 1
  ```
  
Usage:
```
-listen-address string
    	Address to listen on for telemetry (default ":9142")
-telemetry-path string
    	Path under which to expose metrics (default "/metrics")
```
