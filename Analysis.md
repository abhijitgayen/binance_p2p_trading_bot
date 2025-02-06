## Need to find the nearest server

1. api.binance.com

```txt
Non-authoritative answer:
api.binance.com canonical name = d3h36i1mno13q3.cloudfront.net.
Name:   d3h36i1mno13q3.cloudfront.net
Address: 18.164.146.58
```

```json
{
   "status":"success",
   "country":"Singapore",
   "countryCode":"SG",
   "region":"02",
   "regionName":"North East",
   "city":"Singapore",
   "zip":"",
   "lat":1.35538,
   "lon":103.868,
   "timezone":"Asia/Singapore",
   "isp":"Amazon.com, Inc.",
   "org":"AWS CloudFront (GLOBAL)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"18.164.146.58"
}
```

2. api1.binance.com

```txt
Server:         192.168.0.1
Address:        192.168.0.1#53

Non-authoritative answer:
Name:   api1.binance.com
Address: 13.230.189.232
Name:   api1.binance.com
Address: 52.199.153.255
Name:   api1.binance.com
Address: 52.192.245.87
```

```json
{
   "status":"success",
   "country":"Japan",
   "countryCode":"JP",
   "region":"13",
   "regionName":"Tokyo",
   "city":"Tokyo",
   "zip":"150-0002",
   "lat":35.6895,
   "lon":139.692,
   "timezone":"Asia/Tokyo",
   "isp":"Amazon Technologies Inc.",
   "org":"AWS EC2 (ap-northeast-1)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"52.192.245.87"
}
```

3. api2.binance.com

```txt
Server:         192.168.0.1
Address:        192.168.0.1#53

Non-authoritative answer:
Name:   api2.binance.com
Address: 18.177.65.201
Name:   api2.binance.com
Address: 57.180.226.86
```

```json
{
   "status":"success",
   "country":"Japan",
   "countryCode":"JP",
   "region":"13",
   "regionName":"Tokyo",
   "city":"Tokyo",
   "zip":"150-0002",
   "lat":35.6895,
   "lon":139.692,
   "timezone":"Asia/Tokyo",
   "isp":"Amazon Technologies Inc.",
   "org":"AWS EC2 (ap-northeast-1)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"18.177.65.201"
}
```

4. api3.binance.com 

```txt

Server:         192.168.0.1
Address:        192.168.0.1#53

Non-authoritative answer:
Name:   api3.binance.com
Address: 35.72.170.52
Name:   api3.binance.com
Address: 52.193.36.238
```

```json
{
   "status":"success",
   "country":"Japan",
   "countryCode":"JP",
   "region":"13",
   "regionName":"Tokyo",
   "city":"Tokyo",
   "zip":"150-0002",
   "lat":35.6895,
   "lon":139.692,
   "timezone":"Asia/Tokyo",
   "isp":"Amazon Technologies Inc.",
   "org":"AWS EC2 (ap-northeast-1)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"52.193.36.238"
}
```

5. api4.binance.com

```txt
Server:         192.168.0.1
Address:        192.168.0.1#53

Non-authoritative answer:
Name:   api4.binance.com
Address: 54.95.49.154
Name:   api4.binance.com
Address: 54.248.233.54
```

```json
{
   "status":"success",
   "country":"Japan",
   "countryCode":"JP",
   "region":"13",
   "regionName":"Tokyo",
   "city":"Tokyo",
   "zip":"150-0002",
   "lat":35.6895,
   "lon":139.692,
   "timezone":"Asia/Tokyo",
   "isp":"Amazon.com, Inc.",
   "org":"AWS EC2 (ap-northeast-1)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"54.248.233.54"
}
```

6. api5.binance.com

```txt
Server:         192.168.0.1
Address:        192.168.0.1#53

Non-authoritative answer:
Name:   api5.binance.com
Address: 3.114.212.32
Name:   api5.binance.com
Address: 175.41.229.227
Name:   api5.binance.com
Address: 46.51.249.23
```

```json
{
   "status":"success",
   "country":"Japan",
   "countryCode":"JP",
   "region":"13",
   "regionName":"Tokyo",
   "city":"Tokyo",
   "zip":"150-0002",
   "lat":35.6895,
   "lon":139.692,
   "timezone":"Asia/Tokyo",
   "isp":"Amazon.com, Inc.",
   "org":"AWS EC2 (ap-northeast-1)",
   "as":"AS16509 Amazon.com, Inc.",
   "query":"46.51.249.23"
}
```

According to my observation i Am clear that . We need to chose backend api 
one of them

1.  api1.binance.com
2.  api2.binance.com
3.  api3.binance.com
3.  api4.binance.com
4.  api5.binance.com

We need to find the load free server for this also.
