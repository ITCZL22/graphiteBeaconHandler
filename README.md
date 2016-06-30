## graphiteBeaconHandler
### Features
graphite-beacon对于所有的alert只支持全局配置,graphiteBeaconHandler可以对不同的alert支持不同的配置
### Service Config Example
[conf.json](./conf.json)
### Graphite_beacon Config Example 
[graphite_beacon.json](./graphite_beacon.json)
### Prefix--Group Notice
当graphite_beacon多组alert使用相同的通知时，可以利用前缀。相同前缀的alert使用conf.json中由前缀命名的通知组
