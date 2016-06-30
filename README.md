## graphiteBeaconHandler
### Features
    graphite-beacon对于所有alert只支持全局通知,graphiteBeaconHandler可以对不同的alert支持不同的通知
    目前支持的通知方式Email和Slack
### Service Config Example
[conf.json](./conf.json)
### Graphite_beacon Config Example 
[graphite_beacon.json](./graphite_beacon.json)
### Prefix--Group Notice
    当graphite_beacon多组alert使用相同的通知时，可以利用前缀。相同前缀的alert使用conf.json中由前缀命名的通知组
